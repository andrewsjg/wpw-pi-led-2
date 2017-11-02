package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WPTechInnovation/wpw-sdk-go/wpwithin"
	"github.com/WPTechInnovation/wpw-sdk-go/wpwithin/psp"
	"github.com/WPTechInnovation/wpw-sdk-go/wpwithin/psp/onlineworldpay"
	"github.com/WPTechInnovation/wpw-sdk-go/wpwithin/types"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// Application flags
var flagWPServiceKey string
var flagWPClientKey string
var flagIgnoreGPIO bool // Ignore any errors that arise from trying to setup RPi GPIO pins

// Application Vars
var wpw wpwithin.WPWithin
var wpwHandler Handler
var pspConfig map[string]string
var unitsInTime map[int]int

const (
	redDescr   string = "Turn on the red LED"
	greenDescr string = "Turn on the green LED"
	blueDescr  string = "Turn on the blue LED"
	second     string = "second"
	minute     string = "minute"
)

func init() {

	flag.StringVar(&flagWPServiceKey, "wpservicekey", "", "Worldpay service key")
	flag.StringVar(&flagWPClientKey, "wpclientkey", "", "Worldpay client key")
	flag.BoolVar(&flagIgnoreGPIO, "ignoregpio", false, "Ignore GPIO pin errors")
}

func main() {

	err := initLog()
	errCheck(err, "initLog()")

	flag.Parse()

	if strings.EqualFold(flagWPClientKey, "") {
		fmt.Println("Flag wpclientkey is required")
		os.Exit(1)
	} else if strings.EqualFold(flagWPServiceKey, "") {
		fmt.Println("Flag wpservicekey is required")
		os.Exit(1)
	}

	_wpw, err := wpwithin.Initialise("pi-led-producer", "Worldpay Within Pi LED Demo - Producer")
	wpw = _wpw

	errCheck(err, "WorldpayWithin Initialise")

	doSetupServices()
	printProducerOverview()
	fmt.Printf("\n\n")

	// wpwhandler accepts callbacks from worldpay within when service delivery begin/end is required.
	err = wpwHandler.setup(wpw.GetDevice().Services, flagIgnoreGPIO)
	errCheck(err, "wpwHandler setup")
	wpw.SetEventHandler(&wpwHandler)

	err = wpw.InitProducer(pspConfig)
	errCheck(err, "Init producer")
	fmt.Println("Worldpay Within Producer successfully initialised")

	fmt.Println("Starting Service broadcast...")
	err = wpw.StartServiceBroadcast(0) // 0 = no timeout

	errCheck(err, "start service broadcast")

	// run the app until it is closed
	runForever()
}

func doSetupServices() {

	_unitsInTime := make(map[int]int, 0)
	unitsInTime = _unitsInTime

	unitsInTime[1] = 1
	unitsInTime[2] = 60

	////////////////////////////////////////////
	// PSP Configuration
	////////////////////////////////////////////

	_pspConfig := make(map[string]string, 0)
	pspConfig = _pspConfig
	pspConfig[psp.CfgPSPName] = onlineworldpay.PSPName
	pspConfig[onlineworldpay.CfgMerchantClientKey] = flagWPClientKey
	pspConfig[onlineworldpay.CfgMerchantServiceKey] = flagWPServiceKey
	pspConfig[psp.CfgHTEPrivateKey] = flagWPServiceKey
	pspConfig[psp.CfgHTEPublicKey] = flagWPClientKey
	pspConfig[onlineworldpay.CfgAPIEndpoint] = "https://api.worldpay.com/v1"

	////////////////////////////////////////////
	// Red LED
	////////////////////////////////////////////

	svcRedLed, err := types.NewService()
	errCheck(err, "New service - red led")

	svcRedLed.ID = 1
	svcRedLed.Name = "Red LED"
	svcRedLed.Description = redDescr

	priceRedLedSecond, err := types.NewPrice()
	errCheck(err, "Create new price - red led second")

	priceRedLedSecond.Description = redDescr
	priceRedLedSecond.ID = 1
	priceRedLedSecond.UnitDescription = second
	priceRedLedSecond.UnitID = 1
	priceRedLedSecond.PricePerUnit = &types.PricePerUnit{
		Amount:       5,
		CurrencyCode: "GBP",
	}

	svcRedLed.AddPrice(*priceRedLedSecond)

	priceRedLedMinute, err := types.NewPrice()
	errCheck(err, "Create new price - red led minute")

	priceRedLedMinute.Description = redDescr
	priceRedLedMinute.ID = 2
	priceRedLedMinute.UnitDescription = minute
	priceRedLedMinute.UnitID = 2
	priceRedLedMinute.PricePerUnit = &types.PricePerUnit{
		Amount:       20,
		CurrencyCode: "GBP",
	}

	svcRedLed.AddPrice(*priceRedLedMinute)

	err = wpw.AddService(svcRedLed)
	errCheck(err, "Add service - red led")

	////////////////////////////////////////////
	// Green LED
	////////////////////////////////////////////

	svcGreenLed, err := types.NewService()
	errCheck(err, "Create new service - Green LED")
	svcGreenLed.ID = 2
	svcGreenLed.Name = "Green LED"
	svcGreenLed.Description = greenDescr

	priceGreenLedSecond, err := types.NewPrice()
	errCheck(err, "Create new price - green led second")

	priceGreenLedSecond.Description = greenDescr
	priceGreenLedSecond.ID = 1
	priceGreenLedSecond.UnitDescription = second
	priceGreenLedSecond.UnitID = 1
	priceGreenLedSecond.PricePerUnit = &types.PricePerUnit{
		Amount:       10,
		CurrencyCode: "GBP",
	}

	svcGreenLed.AddPrice(*priceGreenLedSecond)

	priceGreenLedMinute, err := types.NewPrice()
	errCheck(err, "Create new price - green led minute")

	priceGreenLedMinute.Description = greenDescr
	priceGreenLedMinute.ID = 2
	priceGreenLedMinute.UnitDescription = minute
	priceGreenLedMinute.UnitID = 2
	priceGreenLedMinute.PricePerUnit = &types.PricePerUnit{
		Amount:       40, /* WOAH! This is minor units so means just 40p */
		CurrencyCode: "GBP",
	}

	svcGreenLed.AddPrice(*priceGreenLedMinute)

	err = wpw.AddService(svcGreenLed)
	errCheck(err, "Add service - green led")

	////////////////////////////////////////////
	// Blue LED
	////////////////////////////////////////////

	svcBlueLed, err := types.NewService()
	errCheck(err, "New service - blue led")

	svcBlueLed.ID = 3
	svcBlueLed.Name = "Blue LED"
	svcBlueLed.Description = redDescr

	priceBlueLedSecond, err := types.NewPrice()
	errCheck(err, "Create new price - blue led second")

	priceBlueLedSecond.Description = blueDescr
	priceBlueLedSecond.ID = 1
	priceBlueLedSecond.UnitDescription = second
	priceBlueLedSecond.UnitID = 1
	priceBlueLedSecond.PricePerUnit = &types.PricePerUnit{
		Amount:       5,
		CurrencyCode: "GBP",
	}

	err = svcBlueLed.AddPrice(*priceBlueLedSecond)
	errCheck(err, "Add service price - blue led second")

	priceBlueLedMinute, err := types.NewPrice()
	errCheck(err, "Create new price - blue led minute")

	priceBlueLedMinute.Description = blueDescr
	priceBlueLedMinute.ID = 2
	priceBlueLedMinute.UnitDescription = minute
	priceBlueLedMinute.UnitID = 2
	priceBlueLedMinute.PricePerUnit = &types.PricePerUnit{
		Amount:       20,
		CurrencyCode: "GBP",
	}

	err = svcBlueLed.AddPrice(*priceBlueLedMinute)
	errCheck(err, "Add service price - blue led minute")

	err = wpw.AddService(svcBlueLed)
	errCheck(err, "Add service - blue led")
}

func errCheck(err error, hint string) {

	if err != nil {
		fmt.Printf("Did encounter error during: %s\n", hint)
		fmt.Println(err.Error())
		fmt.Println("Quitting...")
		os.Exit(1)
	}
}

func runForever() {

	done := make(chan bool)
	fnForever := func() {
		for {
			time.Sleep(time.Second * 10)
		}
	}

	go fnForever()

	<-done // Block forever
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
	}))

	f, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_CREATE, 0755)

	log.SetOutput(f)

	return err
}

func printProducerOverview() {

	device := wpw.GetDevice()

	fmt.Println("Producer Overview:")
	fmt.Println("")
	fmt.Println("Device:")
	fmt.Printf("\tName: %s\n", device.Name)
	fmt.Printf("\tDescription: %s \n", device.Description)
	fmt.Printf("\tIPv4: %s \n", device.IPv4Address)
	fmt.Printf("\tUUID: %s \n", device.UID)
	fmt.Printf("\tServices:\n")
	for _, svc := range device.Services {

		fmt.Printf("\t\tID=%d, Name=%s, Description=%s\n", svc.ID, svc.Name, svc.Description)
		fmt.Printf("\t\t\tPrices: \n")
		for _, price := range svc.Prices {

			fmt.Printf("\t\t\t\tID=%d, Description=%s\n", price.ID, price.Description)
			fmt.Printf("\t\t\t\tUnitID=%d, UnitDescription=%s\n", price.UnitID, price.UnitDescription)
			fmt.Printf("\t\t\t\tCurrency=%s, Amount=%d\n", price.PricePerUnit.CurrencyCode, price.PricePerUnit.Amount)
		}
	}

	fmt.Println("PSP Configuration:")
	for k, v := range pspConfig {

		fmt.Printf("\t%s \t--> %s\n", k, v)
	}
}
