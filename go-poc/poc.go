package main

import (
  "fmt"
  "io"
  //"time"
  "os"
  "os/exec"
  "bufio"
  "github.com/zserge/webview"
)


type ShellCommand struct {
  // Struct for passing around the command for the shell along with the
  // JS callback that handles the result on the GUI.
  command string
  callback string
  multiline bool
}

type Controller struct {
  // No vars here for now,
  // But I believe this should have no state
  // REMEMBER: in JS all will become camelCase (first letter lowercase)
  sharedShellInput chan ShellCommand
  window webview.WebView
}

type fnw func(io.ReadCloser)

// Test function to print on the console
func (c *Controller) SayHello() {
	fmt.Println("[Controller]    Hello! :)");
}

// Return the output of the command line by line
func (c *Controller) GetByLine(command string, callback string) {
	fmt.Println("[Controller]    Sending command to shell: ", command,
    " (callback: ", callback , " - line-by-line)" );
  // Define the output manager
  outputManager := func(bashStdout io.ReadCloser) {
    scanner := bufio.NewScanner(bashStdout)
    for scanner.Scan() {
      SendToJavaScript(callback, scanner.Text(), c.window)
    }
  }
  PrivateShell(command, outputManager)
}

// Returns a single multiline string with the entire output of the command
func (c *Controller) GetAllLines(command string, callback string) {
	fmt.Println("[Controller]    Sending command to shell: ", command,
    " (callback: ", callback , " - all-at-once)" );
  // Define the output manager
  outputManager := func(bashStdout io.ReadCloser) {
    scanner := bufio.NewScanner(bashStdout)
    accumulator := ""
    for scanner.Scan() {
      accumulator += scanner.Text() + "\n"
    }
    SendToJavaScript(callback, accumulator, c.window)
  }
  PrivateShell(command, outputManager)
}

// Calls a JS methodd with one string parameter into the GUI
func SendToJavaScript(method string, parameter string, window webview.WebView){
  fmt.Println("                > ", parameter )
  window.Dispatch(func() {
    window.Eval(fmt.Sprintf("(function(param){ " +
    "  %s(param); " +
    "})(`%s`)", method, parameter ))
  })
}

// Manages a private shell instance where to run a single command into.
func PrivateShell(input string, outputManager fnw) {
  // Init the shell
  fmt.Println("[Shell routine] Starting up private shell...")
  bashProcess := exec.Command("bash", "-c", input)
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
  // Prepare output reader
  fmt.Println("[Shell routine] Setup output manager...")
  go outputManager(bashStdout) // This function comes as a parameter!

  fmt.Println("[Shell routine] Run!")
  go bashProcess.Run()
}


// Sends a command to the shared shell, discarding any output
func (c *Controller) Send(command string) {
	fmt.Println("[Controller]    Sending command to shell: ", command,
    "(no callback)" );
  c.sharedShellInput <- ShellCommand{command, "", false}
}

// Manages the shared shell instance where to run the commands into.
// To avoid multiple wasted initializations, it keeps running and waits
// for commands through its own *ShellCommand channel
func SharedShell(input<- chan ShellCommand, quit<- chan bool, w webview.WebView) {

  // Init the shell
  fmt.Println("[Shell routine] Starting up shared shell...")
  bashProcess := exec.Command("bash")
  bashStdin, errin := bashProcess.StdinPipe()
  if errin != nil {
      fmt.Fprintln(os.Stderr, "Error creating StdinPipe!", errin)
      return
  }
  err := bashProcess.Start()
  if err != nil {
      fmt.Fprintln(os.Stderr, "Error starting bash", err)
      return
  }
  fmt.Println("[Shell routine] Shell setup complete.")

  // Main loop
  for {
    select {
    case i := <-input:
      fmt.Println("[Shell routine] Executing command: ", i.command,
                                          "  (callback: ", i.callback, ")")
      bashStdin.Write([]byte( i.command ))
      bashStdin.Write([]byte("&\n"))
      // & to immediately release bash's stdout
      // \n to send the command
    case <-quit:
      break ;
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

    // Start up the goroutine managing the shared shell
    quitChannel := make(chan bool)
    controller.sharedShellInput = make(chan ShellCommand)
    controller.window = w

    go SharedShell(controller.sharedShellInput, quitChannel, w)
    controller.Send("echo 'Shell ready'")

    // Inject controller
    fmt.Println("[Main]          Injecting controller into the interface...")
		w.Bind("controller", controller)

    // Done!
    fmt.Println("[Main]          Setup complete. Entering main loop...")
  });

  w.Run()
}
