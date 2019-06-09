#define WEBVIEW_IMPLEMENTATION
#include <stdio.h>
#include <stdlib.h>
#include "webview.h"
#include "jsmn.h"

void my_cb(struct webview *w, const char *arg);

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
  
  /* Main app loop, can be either blocking or non-blocking */
  while (webview_loop(&webview, 1) == 0);
  webview_exit(&webview);

  return 0;
}

// JS "invoke" callback
void my_cb(struct webview *w, const char *arg) {
	printf("Call received! Let me read this: %s\n", arg);
	
	jsmn_parser jsmn_parser;
	printf("- parser instantiated...\n");
    jsmntok_t tokens[1000]; /* We expect no more than 20 JSON tokens */
	printf("- 20 tokens allocated...\n");

    jsmn_init(&jsmn_parser);
    printf("- parser initialized...\n");

    int result = jsmn_parse(&jsmn_parser, arg, strlen(arg), tokens, 128);
    printf("- parsing complete.\n");

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
    
    for(int i=1; i<result; i=i+2){
        // Get the pair 
        char typeof_command[20]; // Command types can be maximum 20 chars long;
        memcpy( typeof_command, &arg[tokens[i].start], tokens[i].end-tokens[i].start );
        typeof_command[tokens[i].end-tokens[i].start] = '\0';
        char actual_command[1000]; // Command types can be maximum 1000 chars long;
        memcpy( actual_command, &arg[tokens[i+1].start], tokens[i+1].end-tokens[i+1].start );
        actual_command[tokens[i+1].end-tokens[i+1].start] = '\0';


        printf("#### %s %s\n", typeof_command, actual_command);
        if(!strcmp(typeof_command, "send_command")){
            printf("- Command to be sent: %s  -> Sending it!\n\n", actual_command);
            system(actual_command);
            printf("\n  Done\n");
            
        }
        if(!strcmp(typeof_command, "send_and_read")){
            printf("- Command to be sent and read back: %s  -> Sending it!\n\n", actual_command);
            
              FILE *fp;
              char path[1035];
              
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
              while (fgets(path, sizeof(path)-1, fp) != NULL) {
                i++;
                printf("Output line %i of '%s': %s", i, actual_command, path);
              }
              /* close */
              pclose(fp);
            
            
            printf("\n  Done\n");
        }

    }
    
}



