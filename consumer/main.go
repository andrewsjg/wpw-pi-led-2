package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/wptechinnovation/wpw-sdk-go/wpwithin"
	"github.com/wptechinnovation/wpw-sdk-go/wpwithin/psp"
	"github.com/wptechinnovation/wpw-sdk-go/wpwithin/psp/onlineworldpay"
	wpwtypes "github.com/wptechinnovation/wpw-sdk-go/wpwithin/types"
)

// Application flags
var flagProducerUUID string
var flagServiceID int
var flagPriceID int
var flagUnitQuantity int
var flagDiscoveryTimeout int
var flagInteractive bool

// Application Vars
var wpw wpwithin.WPWithin
var hceCard *wpwtypes.HCECard

func init() {

	flag.StringVar(&flagProducerUUID, "produceruuid", "", "Producer UUID")
	flag.IntVar(&flagServiceID, "serviceid", 1, "Service ID")
	flag.IntVar(&flagPriceID, "priceid", 1, "Price ID")
	flag.IntVar(&flagUnitQuantity, "unitquantity", 2, "Unit quantity")
	flag.IntVar(&flagDiscoveryTimeout, "discoverytimeout", 20000, "Device discovery timeout (millis)")
	flag.BoolVar(&flagInteractive, "interactive", false, "Interactive mode - prompt for carriage return between steps")
}

func main() {

	err := initLog()
	errCheck(err, "initLog()")

	flag.Parse()

	if strings.EqualFold(flagProducerUUID, "") {

		fmt.Println("Producer UUID is not set")
		fmt.Println("Please specify -produceruuid <....>")
		os.Exit(1)
	}

	err = performSetup()
	errCheck(err, "performSetup()")

	_wpw, err := wpwithin.Initialise("pi-led-consumer", "Worldpay Within Pi LED Demo - Consumer", "")
	wpw = _wpw

	errCheck(err, "WorldpayWithin Initialise")

	printConsumerOverview()
	promptContinue()
	doConsumeService()
}

func doConsumeService() {

	// Device discovery
	fmt.Printf("Performing device discovery with timeout %dms\n", flagDiscoveryTimeout)
	bm, err := wpw.DeviceDiscovery(flagDiscoveryTimeout)
	errCheck(err, "wpw.DeviceDiscovery()")

	fmt.Printf("Found %d devices, filtering on device with UUID = %s\n", len(bm), flagProducerUUID)

	var selectedBM *wpwtypes.BroadcastMessage
	for _, bm := range bm {

		if strings.EqualFold(bm.ServerID, flagProducerUUID) {

			fmt.Printf("Found required device %s - %s\n", bm.DeviceDescription, bm.ServerID)

			selectedBM = &bm
			break
		}
	}

	if selectedBM == nil {

		fmt.Printf("Specified producer not found (%s)\n", flagProducerUUID)
		os.Exit(1)
	}

	var pspConfig = make(map[string]string, 0)
	pspConfig[psp.CfgPSPName] = onlineworldpay.PSPName
	pspConfig[onlineworldpay.CfgAPIEndpoint] = "https://api.worldpay.com/v1"

	fmt.Printf("Setting up connection with %s\n", selectedBM.DeviceDescription)
	fmt.Printf("\n\n")

	err = wpw.InitConsumer(selectedBM.Scheme, selectedBM.Hostname, selectedBM.PortNumber, selectedBM.URLPrefix, "123", hceCard, pspConfig)
	errCheck(err, "wpw.InitConsumer()")
	fmt.Println("Requesting services..")
	// Service discovery
	svcs, err := wpw.RequestServices()
	errCheck(err, "wpw.RequestServices()")

	var selectedSVC *wpwtypes.ServiceDetails
	for _, svc := range svcs {

		if svc.ServiceID == flagServiceID {

			fmt.Printf("Found required service %d - %s\n", flagServiceID, svc.ServiceName)
			selectedSVC = &svc
			break
		}
	}

	if selectedSVC == nil {

		fmt.Printf("Specified service not found (%d)\n", flagServiceID)
		os.Exit(1)
	}

	fmt.Printf("\n\n")

	// Price discovery
	fmt.Println("Requesting service prices..")
	svcPrices, err := wpw.GetServicePrices(selectedSVC.ServiceID)
	errCheck(err, "wpw.GetServicePrices()")

	var selectedPrice *wpwtypes.Price
	for _, price := range svcPrices {

		if price.ID == flagPriceID {

			fmt.Printf("Found required price %d - %s @%s %dp per %s\n", flagServiceID, price.Description, price.PricePerUnit.CurrencyCode, price.PricePerUnit.Amount, price.UnitDescription)
			selectedPrice = &price
			break
		}
	}

	if selectedPrice == nil {

		fmt.Printf("Specified price not found (%d)\n", flagPriceID)
		os.Exit(1)
	}

	promptContinue()
	fmt.Printf("\n\n")

	// Service + price selection
	fmt.Println("Selecting service and price.. Getting quote for:")
	fmt.Printf("%s - %d units of %s @ %s %dp per unit\n", selectedPrice.Description, flagUnitQuantity, selectedPrice.UnitDescription, selectedPrice.PricePerUnit.CurrencyCode, selectedPrice.PricePerUnit.Amount)
	fmt.Println()
	totalPriceResponse, err := wpw.SelectService(selectedSVC.ServiceID, flagUnitQuantity, selectedPrice.ID)
	errCheck(err, "wpw.SelectService()")

	fmt.Println("TotalPriceResponse:")
	fmt.Printf("Total price %dp\n", totalPriceResponse.TotalPrice)
	fmt.Printf("Merchant Public Key: %s\n", totalPriceResponse.MerchantClientKey)
	fmt.Printf("Currency: %s\n", totalPriceResponse.CurrencyCode)
	fmt.Printf("Reference: %s\n", totalPriceResponse.PaymentReferenceID)
	fmt.Printf("Units to supply: %d\n", totalPriceResponse.UnitsToSupply)

	promptContinue()
	fmt.Printf("\n\n")

	// Payment request
	fmt.Printf("Proceed to make payment of %dp\n", totalPriceResponse.TotalPrice)
	fmt.Printf("Payment card for %s %s, number %s, with expiry %d/%d\n", hceCard.FirstName, hceCard.LastName, hceCard.CardNumber, hceCard.ExpMonth, hceCard.ExpYear)
	paymentResponse, err := wpw.MakePayment(totalPriceResponse)
	errCheck(err, "wpw.MakePayment()")

	fmt.Println("Worldpay Within payment successful")

	fmt.Printf("\n\n")

	fmt.Println("PaymentResponse:")
	fmt.Printf("Total paid: %dp\n", paymentResponse.TotalPaid)
	fmt.Printf("DeliveryToken - Key: %s\n", paymentResponse.ServiceDeliveryToken.Key)
	fmt.Printf("DeliveryToken - Issued: %s\n", paymentResponse.ServiceDeliveryToken.Issued)
	fmt.Printf("DeliveryToken - Expiry: %s\n", paymentResponse.ServiceDeliveryToken.Expiry)
	fmt.Printf("DeliveryToken - Refund on expiry: %t\n", paymentResponse.ServiceDeliveryToken.RefundOnExpiry)

	promptContinue()
	fmt.Printf("\n\n")

	// Begin service delivery

	fmt.Println("Proceed to begin service delivery (Turn on the LED)")

	promptContinue()
	fmt.Printf("\n\n")

	_, err = wpw.BeginServiceDelivery(selectedSVC.ServiceID, *paymentResponse.ServiceDeliveryToken, flagUnitQuantity)
	errCheck(err, "wpw.BeginServiceDelivery()")
	fmt.Printf("\n\n")
	fmt.Printf("%s should be powered on for %d * %s\n", selectedSVC.ServiceName, flagUnitQuantity, selectedPrice.UnitDescription)
	fmt.Printf("\n\n")
}

func performSetup() error {

	hceCard = &wpwtypes.HCECard{
		FirstName:  "John",
		LastName:   "Smith",
		ExpMonth:   5,
		ExpYear:    2019,
		CardNumber: "4444333322221111",
		Type:       "Card",
		Cvc:        "123",
	}

	return nil
}

func errCheck(err error, hint string) {

	if err != nil {
		fmt.Printf("Did encounter error during: %s\n", hint)
		fmt.Println(err.Error())
		fmt.Println("Quitting...")
		os.Exit(1)
	}
}

func initLog() error {

	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

	log.AddHook(lfshook.NewHook(lfshook.PathMap{
		log.InfoLevel:  "logs/info.log",
		log.ErrorLevel: "logs/error.log",
		log.DebugLevel: "logs/debug.log",
		log.WarnLevel:  "logs/warn.log",
		log.PanicLevel: "logs/panic.log",
		log.FatalLevel: "logs/fatal.log",
	}, new(log.TextFormatter)))

	f, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_CREATE, 0755)

	log.SetOutput(f)

	return err
}

func printConsumerOverview() {

	fmt.Printf("Device discovery timeout: %dms\n", flagDiscoveryTimeout)
	fmt.Printf("Device UUID filter: %s\n", flagProducerUUID)
	fmt.Printf("Service ID filter: %d\n", flagServiceID)
	fmt.Printf("Price ID filter %d\n", flagPriceID)
	fmt.Printf("Order quantity: %d\n", flagUnitQuantity)

	fmt.Printf("------------------------------------------\n\n\n")
}

func promptContinue() {

	if flagInteractive {

		fmt.Println("<return to continue>")
		fmt.Scanf("\n", nil)
	}
}
