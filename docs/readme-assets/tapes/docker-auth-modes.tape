Output docs/readme-assets/docker-auth-modes.gif
Set Shell bash
Set FontSize 22
Set Width 1280
Set Height 760
Set Padding 24
Set Margin 20
Set MarginFill "#efe7dc"
Set BorderRadius 12
Set WindowBar Rings
Set CursorBlink false
Set TypingSpeed 35ms
Set Theme {"black":"#24313d","red":"#cb4b5a","green":"#2f8f6b","yellow":"#b07d1a","blue":"#3c78d8","magenta":"#8f4ec6","cyan":"#2a7f8a","white":"#f5f0e6","brightBlack":"#51606d","brightRed":"#dd6b75","brightGreen":"#52a884","brightYellow":"#c8942c","brightBlue":"#5d93e3","brightMagenta":"#a06cd5","brightCyan":"#4e9ca6","brightWhite":"#fffaf0","background":"#18222d","foreground":"#f5f0e6","selection":"#304150","cursor":"#f5f0e6"}

Type "./bin/faultline analyze examples/docker-auth.log"
Enter
Wait+Screen /Summary/
Sleep 1400ms
Type "clear"
Enter
Sleep 500ms
Type "./bin/faultline analyze examples/docker-auth.log --mode detailed"
Enter
Wait+Screen /TRIGGERED BY/
Sleep 1600ms
Screenshot docs/readme-assets/docker-auth-modes.png
Sleep 1800ms