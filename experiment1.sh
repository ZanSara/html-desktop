
./webview-example&
/bin/sleep 0.3

#wmctrl -r e182d4d56ea0fe8601cc65486e757ebf -b add,fullscreen
#wmctrl -r e182d4d56ea0fe8601cc65486e757ebf -b add,below
wmctrl -r e182d4d56ea0fe8601cc65486e757ebf -b add,skip_taskbar

#xdotool search --name e182d4d56ea0fe8601cc65486e757ebf behave %@ focus exec ./wmctrl-stuff.sh
