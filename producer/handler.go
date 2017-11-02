package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/WPTechInnovation/wpw-sdk-go/wpwithin/types"
	"github.com/stianeikeland/go-rpio"
)

// Handler handles the events coming from Worldpay Within
type Handler struct {
	ledGreen    rpio.Pin
	ledRed      rpio.Pin
	ledBlue     rpio.Pin
	services    map[int]*types.Service
	rpioenabled bool
}

func (handler *Handler) setup(services map[int]*types.Service, ignoreGPIO bool) error {

	if services == nil {

		return errors.New("Services must be set.")
	}

	handler.services = services
	handler.ledRed = rpio.Pin(2)
	handler.ledGreen = rpio.Pin(3)
	handler.ledBlue = rpio.Pin(4)

	gpioErr := rpio.Open()

	if gpioErr != nil {
		fmt.Println("Failed to open Raspberry Pi GPIO")

		if !ignoreGPIO {

			return gpioErr
		}

		fmt.Println("Ignore Raspberry Pi GPIO errors")
		return nil
	}

	// Did successfully setup rpio

	handler.rpioenabled = true

	fmt.Println("Did open Raspberry Pi GPIO")

	// Cleanup (defer until end)
	// rpio.Close()

	// Ensure pins are in output mode
	handler.ledGreen.Output()
	handler.ledRed.Output()
	handler.ledBlue.Output()
	fmt.Println("Did set GPIO pins to output type")

	// Turn of both LEDs, set the pins to low.
	handler.ledGreen.Low()
	handler.ledRed.Low()
	handler.ledBlue.Low()
	fmt.Println("Did set GPIO pins to low")

	return nil
}

// BeginServiceDelivery is called by Worldpay Within when a consumer wish to begin delivery of a service
func (handler *Handler) BeginServiceDelivery(serviceID int, servicePriceID int, serviceDeliveryToken types.ServiceDeliveryToken, unitsToSupply int) {

	fmt.Printf("BeginServiceDelivery. ServiceID = %d\n", serviceID)
	fmt.Printf("BeginServiceDelivery. ServicePriceID = %d\n", servicePriceID)
	fmt.Printf("BeginServiceDelivery. UnitsToSupply = %d\n", unitsToSupply)
	fmt.Printf("BeginServiceDelivery. DeliveryToken = %+v\n", serviceDeliveryToken.Key)
	fmt.Println()
	svc := handler.services[serviceID]

	if &svc == nil {

		fmt.Printf("Service %d not found", serviceID)
		return
	}

	price := svc.Prices[1]

	durationSeconds := unitsToSupply * (unitsInTime[price.ID])
	fmt.Println("Warning, hardcoded price selection due to WPW design flaw. i.e. This event doesn't know what price was selected..")
	fmt.Printf("(%d) %s -> %s for %d %s\n", svc.ID, svc.Name, price.Description, durationSeconds, price.UnitDescription)

	var selectedPin rpio.Pin

	fmt.Print("POWER ON ")
	switch svc.ID {

	case 1:
		fmt.Println("RED LED")
		selectedPin = handler.ledRed
	case 2:
		fmt.Println("GREEN LED")
		selectedPin = handler.ledGreen
	case 3:
		fmt.Println("BLUE LED")
		selectedPin = handler.ledBlue
	default:
		fmt.Println("Unknown service id")
	}

	if handler.rpioenabled {

		selectedPin.High()
	} else {

		fmt.Println("Raspberry Pi GPIO disabled.")
	}

	time.Sleep(time.Duration(durationSeconds) * time.Second)

	fmt.Println("Time is up.. calling EndServiceDelivery()..")
	fmt.Println()

	handler.EndServiceDelivery(serviceID, serviceDeliveryToken, unitsToSupply)
}

// EndServiceDelivery is called by Worldpay Within when a consumer wish to end delivery of a service
func (handler *Handler) EndServiceDelivery(serviceID int, serviceDeliveryToken types.ServiceDeliveryToken, unitsReceived int) {

	fmt.Printf("EndServiceDelivery. ServiceID = %d\n", serviceID)
	fmt.Printf("EndServiceDelivery. UnitsReceived = %d\n", unitsReceived)
	fmt.Printf("EndServiceDelivery. DeliveryToken = %+v\n", serviceDeliveryToken.Key)
	fmt.Println()
	svc := handler.services[serviceID]

	if &svc == nil {

		fmt.Printf("Service %d not found", serviceID)
		return
	}

	fmt.Printf("%d - %s\n", svc.ID, svc.Name)

	var selectedPin rpio.Pin

	fmt.Print("POWER OFF ")
	switch svc.ID {

	case 1:
		fmt.Println("RED LED")
		selectedPin = handler.ledRed
	case 2:
		fmt.Println("GREEN LED")
		selectedPin = handler.ledGreen
	case 3:
		fmt.Println("BLUE LED")
		selectedPin = handler.ledBlue
	default:
		fmt.Println("Unknown service id")
	}

	if handler.rpioenabled {

		selectedPin.Low()
	} else {

		fmt.Println("Raspberry Pi GPIO disabled.")
	}
}

// GenericEvent handles general events
func (handler *Handler) GenericEvent(name string, message string, data interface{}) error {

	return nil
}

func (handler *Handler) MakePaymentEvent(totalPrice int, orderCurrency string, clientToken string, orderDescription string, uuid string) {

	fmt.Printf("go event from core - payment: totalPrice=%d, orderCurrency=%s, clientToken=%s, orderDescription=%s, uui=%s\n",
		totalPrice, orderCurrency, clientToken, orderDescription, uuid)
}

func (handler *Handler) ServiceDiscoveryEvent(remoteAddr string) {

	fmt.Printf("go event from core - service dicovery: remoteAddr: %s\n", remoteAddr)
}

func (handler *Handler) ServicePricesEvent(remoteAddr string, serviceId int) {

	fmt.Printf("go event from core - service prices: remoteAddr: %s, serviceId: %d\n", remoteAddr, serviceId)
}

func (handler *Handler) ServiceTotalPriceEvent(remoteAddr string, serviceId int, totalPrice *types.TotalPriceResponse) {

	fmt.Printf("go event from core - service prices: remoteAddr: %s, serviceId: %d\n", remoteAddr, serviceId)
}

// GenericEvent handles general events
func (handler *Handler) ErrorEvent(msg string) {

	fmt.Printf("go event from core - ErrorEvent: %s\n", msg)
}
