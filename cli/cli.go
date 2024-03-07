package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/MrDweller/technician/technician"
)

type Cli struct {
	technician *technician.Technician
	running    bool
}

type TemperatureResponse struct {
	Temperature float64 `json:"temperature"`
}

func StartCli(technician *technician.Technician) {

	var output io.Writer = os.Stdout
	var input *os.File = os.Stdin

	fmt.Fprintln(output, "Starting sensor retrieval system cli...")

	cli := Cli{
		technician: technician,
		running:    true,
	}

	for {
		if !cli.running {
			fmt.Fprintln(output, "Stopping the sensor retrieval system!")

			err := technician.StopTechnician()
			if err != nil {
				log.Panic(err)
			}
			break
		}

		fmt.Fprint(output, "enter command: ")

		reader := bufio.NewReader(input)
		input, _ := reader.ReadString('\n')

		commands := strings.Fields(input)
		cli.handleCommand(output, commands)
	}
}

func (cli *Cli) Stop() {
	cli.running = false
}

func (cli *Cli) handleCommand(output io.Writer, commands []string) {
	numArgs := len(commands)
	if numArgs <= 0 {
		fmt.Fprintln(output, errors.New("no command found"))
		return
	}

	command := strings.ToLower(commands[0])

	switch command {
	case "subscribe":
		if numArgs == 2 {
			event := commands[1]

			err := cli.technician.Subscribe(event)

			if err != nil {
				fmt.Fprintln(output, err)
			}
		}

	case "unsubscribe":
		if numArgs == 2 {
			event := commands[1]

			err := cli.technician.Unsubscribe(event)

			if err != nil {
				fmt.Fprintln(output, err)
			}
		}

	case "help":
		fmt.Fprintln(output, helpText)

	case "clear":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

	case "exit":
		cli.Stop()

	default:
		fmt.Fprintln(output, errors.New("no command found"))
	}

}

var helpText = `
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	[ SENSOR RETRIEVAL APPLICATION SYSTEM COMMAND LINE INTERFACE ]
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

COMMANDS:
	command [command options] [args...]

VERSION:
	v1.0
	
COMMANDS:
	subscribe <event>			Subscribe to a specifed event
	unsubscribe <event>			Unsubscribe from a specifed event
	help						Output this help prompt
	clear						Clear the terminal
	exit						Stop the sensor retrieval system
`
