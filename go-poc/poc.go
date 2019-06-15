package main

import (
  "fmt"
  //"strings"
  "os"
  "os/exec"
  "bufio"
  //"io/ioutil"
  "./webview"
)


type Controller struct {
  // No vars here for now,
  // But I believe this should have no state
  // REMEMBER: in JS all will become camelCase (first letter lowercase)
  shellInputChannel chan string
}

func (c *Controller) SayHello() {
	fmt.Println("[Controller]    Hello! :)");
}

func (c *Controller) ExecCommand(command string) {
	fmt.Println("[Controller]    Sending command to shell: ", command );
  c.shellInputChannel <- command
}



func shell(input chan string, quit chan string) {
  fmt.Println("[Shell routine] Starting up shell...")

  bashProcess := exec.Command("bash")

  bashStdin, errin := bashProcess.StdinPipe()
  if errin != nil {
      fmt.Fprintln(os.Stderr, "Error creating StdinPipe!", errin)
      return
  }
  bashStdout, errout := bashProcess.StdoutPipe()
  if errout != nil {
      fmt.Fprintln(os.Stderr, "Error creating StdoutPipe!", errout)
      return
  }

  err := bashProcess.Start()
  if err != nil {
      fmt.Fprintln(os.Stderr, "Error starting bash", err)
      return
  }

  fmt.Println("[Shell routine] Shell setup complete.")

  for {
    select {
    case command := <-input:
      fmt.Println("[Shell routine] Executing command: ", command)

      bashStdin.Write([]byte( command ))
      bashStdin.Write([]byte("&\n"))  // & to immediately release bash's stdout
                                      // \n to send the command
      // Create a scanner that reads the output one line at time
      scanner := bufio.NewScanner(bashStdout)
      // Read the stdout line by line async, letting the main thread go
      go func() {
          for scanner.Scan() {
              fmt.Printf("                > %s\n", scanner.Text())
          }
      }()

    case <-quit:
      fmt.Println("[Shell Routine] Quit signal received. Goodbye!")
      break;

    }
  }

  bashStdin.Close()
  bashProcess.Wait() // Is this any useful?
}

func main() {
	w := webview.New(webview.Settings{
		Title: "justatentative",
		URL:   "file:///home/s/Projects/html-desktop/go-poc/myindex.html",
		Resizable: true,
	})
	w.SetColor(255, 255, 255, 0)

  defer w.Exit()

  w.Dispatch(func() {
    // Create controller
    controller := new(Controller)

    // Start up the goroutine managing the shell
    quitChannel := make(chan string, 1 )
    controller.shellInputChannel = make(chan string)
    go shell(controller.shellInputChannel, quitChannel)
    controller.ExecCommand("echo 'Shell ready'")

    // Inject controller
    fmt.Println("[Main]          Injecting controller into the interface...")
		w.Bind("controller", controller)

    // Done!
    fmt.Println("[Main]          Setup complete. Entering main loop...")
  });

  w.Run()
}
