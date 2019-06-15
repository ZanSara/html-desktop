#define WEBVIEW_IMPLEMENTATION
#include <stdio.h>
#include <stdlib.h>
#include <dbus/dbus.h>

// http://www.microhowto.info/howto/capture_the_output_of_a_child_process_in_c.html
#include <errno.h> 	//errno, EINTR
#include <stdio.h> 	//perror
#include <stdlib.h> //exit
#include <unistd.h> //_exit, close, dup2, execl, fork, pipe, STDOUT_FILENO
#include <sys/wait.h> 	// wait, pid_t
#include <fcntl.h> 	//fnctl, F_SETFL, O_NONBLOCK

#include "webview.h"
#include "jsmn.h"

void my_cb(struct webview *w, const char *arg);
void monitor_dbus_events(const char* interface_name);

int main() {
  printf("Starting upp!\n");
  struct webview webview = {
      .title = "e182d4d56ea0fe8601cc65486e757ebf",
      .url =  "file:///home/s/Projects/C/myindex.html",
      .width = 800,
      .height = 600,
      .debug = 1,
      .resizable = 1,
      
  };
  webview.external_invoke_cb = my_cb;
  webview_init(&webview);
  webview_set_color(&webview, 255, 255, 255, 0);
      
  //monitor_dbus_events("backlight");
    
  /* Main app loop, can be either blocking or non-blocking */
  while (webview_loop(&webview, 1) == 0);
  webview_exit(&webview);
  return 0;
}

// JS "invoke" callback
void my_cb(struct webview *w, const char *arg) {
	printf("Call received! Let me read this: %s\n", arg);
	
	jsmn_parser jsmn_parser;
	jsmntok_t tokens[1000]; /* We expect no more than 20 JSON tokens */
	jsmn_init(&jsmn_parser);
    int result = jsmn_parse(&jsmn_parser, arg, strlen(arg), tokens, 128);
    
    /*
    printf("- Read %i tokens:\n", result);
    for(int i=0; i<result; i++){
        if(tokens[i].type == 0){
            printf("    undefined type o.O\n");
            continue;
        }         
        if(tokens[i].type  == 1){
            printf("    Object type: %.*s (size: %i)\n", tokens[i].end-tokens[i].start, arg + tokens[i].start, tokens[i].end-tokens[i].start );
            continue;
        } 
        if(tokens[i].type  == 2){
            printf("    Array type: %.*s (size: %i)\n", tokens[i].end-tokens[i].start, arg + tokens[i].start, tokens[i].end-tokens[i].start );
            continue;
        } 
        if(tokens[i].type  == 3){
            printf("    String type: %.*s (size: %i)\n", tokens[i].end-tokens[i].start, arg + tokens[i].start, tokens[i].end-tokens[i].start );
            continue;
        }
        if(tokens[i].type  == 4){
            printf("    Primitive type: %.*s (size: %i)\n", tokens[i].end-tokens[i].start, arg + tokens[i].start, tokens[i].end-tokens[i].start );
            continue;
        } else {
            printf("    Unknown type!! Code %i\n", tokens[i].type );
        }
    }
    printf("Ok! Now let's understand it!\n");
    */
    
    for(int i=1; i<result; i=i+2){
        // Get the pair 
        char typeof_command[20]; // Command types can be maximum 20 chars long;
        memcpy( typeof_command, &arg[tokens[i].start], tokens[i].end-tokens[i].start );
        typeof_command[tokens[i].end-tokens[i].start] = '\0';
        char actual_command[1000]; // Command types can be maximum 1000 chars long;
        memcpy( actual_command, &arg[tokens[i+1].start], tokens[i+1].end-tokens[i+1].start );
        actual_command[tokens[i+1].end-tokens[i+1].start] = '\0';


        // printf("#### %s %s\n", typeof_command, actual_command);
        if(strcmp(typeof_command, "send_command") == 0){
            printf("- Command to be sent: %s  -> Sending it!\n\n", actual_command);
            system(actual_command);
            printf("\n  Done\n");
            
        }
        if(strcmp(typeof_command, "send_and_read") == 0){
            printf("- Command to be sent and read back: %s  -> Sending it!\n\n", actual_command);
            
              FILE *fp;
              char line[1035];
              
              char fullpath[1500] = {0};
              snprintf(fullpath, sizeof(fullpath), "%s/%s ", "/bin", actual_command);

              /* Open the command for reading. */
              fp = popen(fullpath, "r");
              if (fp == NULL) {
                printf("Failed to run command '%s'\n", actual_command );
                exit(1);
              }
              /* Read the output a line at a time - output it. */
              int i = 0;
              while (fgets(line, sizeof(line)-1, fp) != NULL) {
                i++;
                printf("Output line %i of '%s': %s", i, actual_command, line);
              }
              /* close */
              pclose(fp);
            
            
            printf("\n  Done\n");
        }

    }
    
}



/**
* Run in the background
*/
void monitor_dbus_events(const char* interface_name )
{
    // Create pipe
    int filedes[2];
    if (pipe(filedes) == -1) {
        perror("pipe");
        exit(1);
    }
    
    // Set non-blocking on the readable end.
    if (fcntl(filedes[0], F_SETFL, O_NONBLOCK))
    {
        close(filedes[0]);
        close(filedes[1]);
        return -1;
    }
    
    // Fork
    int pid = fork();
    if(pid == -1)
    {
        printf("Fork Failed\n");
        return -1;
    }
    if(pid == 0 )
    {
        // Only child process continues
        while ((dup2(filedes[1], STDOUT_FILENO) == -1) && (errno == EINTR)) {}
        close(filedes[1]);
        close(filedes[0]);
        execl("/bin/dbus-monitor", "dbus-monitor --system", (char*)0);
        perror("execl");
        _exit(1);
    } 
    else
    {
        char buffer[4096];
        while (1) 
        {
            ssize_t count = read(filedes[0], buffer, sizeof(buffer));
            if (count == -1) 
            {
                if (errno == EINTR) 
                {
                    continue;
                } 
                else 
                {
                    perror("read");
                    exit(1);
                }
            } 
            else if (count == 0) 
            {
                break;
            } 
            else 
            {
                //handle_child_process_output(buffer, count);
                printf("######################## Now handle child process output! %s\n", buffer);
            }
        }
        close(filedes[0]);
        wait(0);
    }    
    // Both parent & child closes
    close(filedes[1]);
    return 0;
}

 /*   
    // Create pipe
    int filedes[2];
    if (pipe(filedes) == -1) {
      perror("pipe");
      exit(1);
    }
    
    int pid = fork();
    if(pid == -1){
        printf("Fork Failed\n");
        return;
    }
    if(pid == 0 ){
        // Only child process continues
        while ((dup2(filedes[1], STDOUT_FILENO) == -1) && (errno == EINTR)) {}
        close(filedes[1]);
        close(filedes[0]);
        execl("/bin/dbus-monitor", "dbus-monitor --system", (char*)0);
        perror("execl");
        _exit(1);

    } else {
    
        char buffer[4096];
        while (1) {
          ssize_t count = read(filedes[0], buffer, sizeof(buffer));
          if (count == -1) {
            if (errno == EINTR) {
              continue;
            } else {
              perror("read");
              exit(1);
            }
          } else if (count == 0) {
            break;
          } else {
            //handle_child_process_output(buffer, count);
            printf("######################## Now handle child process output! %s\n", buffer);
          }
        }
        close(filedes[0]);
        wait(0);
        
    }    
    // Both parent & child closes
    close(filedes[1]);

}

void handle_dbus_monitor_messages();



/*
    FILE *fp;
    char line[1035];

    char fullpath[1500] = {0};

    /* Open the command for reading. * /
    fp = popen("/bin/dbus-monitor --system --monitor", "r");
    if (fp == NULL) {
        printf("Failed to launch dbus-monitor D:\n" );
        exit(1);
    }
    /* Read the output a line at a time * /
    int i = 0;
    int accumulating = 0;
    char* acc_signal = "";
    while (fgets(line, sizeof(line)-1, fp) != NULL) {
        i++;
        if(strncmp("signal", line, 6) == 0) { // if line starts with "signal"
            
            acc_signal = "";
            // See what comes next
            if(strstr(line, interface_name)){
                accumulating = 1; // Print until the next signal shows up
            } else {
                accumulating = 0; // Don't print until the next signal shows up
            }
        } else {
            if(accumulating){
                printf(line);
                strcat(acc_signal, line);
            } 
        }
    }
    /* close * /
    pclose(fp);
    
    */
//}


