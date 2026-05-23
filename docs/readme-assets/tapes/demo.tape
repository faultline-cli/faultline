Output docs/readme-assets/demo.gif
Set Shell bash
Set FontSize 22
Set Width 1280
Set Height 1200
Set Padding 24
Set Margin 20
Set MarginFill "#efe7dc"
Set BorderRadius 12
Set WindowBar Rings
Set CursorBlink false
Set TypingSpeed 40ms
Set Theme {"black":"#24313d","red":"#cb4b5a","green":"#2f8f6b","yellow":"#b07d1a","blue":"#3c78d8","magenta":"#8f4ec6","cyan":"#2a7f8a","white":"#f5f0e6","brightBlack":"#51606d","brightRed":"#dd6b75","brightGreen":"#52a884","brightYellow":"#c8942c","brightBlue":"#5d93e3","brightMagenta":"#a06cd5","brightCyan":"#4e9ca6","brightWhite":"#fffaf0","background":"#18222d","foreground":"#f5f0e6","selection":"#304150","cursor":"#f5f0e6"}

# Step 1 — classify a failing log
Type "./bin/faultline analyze examples/runtime-mismatch.log"
Enter
Wait+Screen /faultline workflow/
Sleep 1800ms
Screenshot docs/readme-assets/demo.png
Sleep 1600ms

# Step 2 — get the fix steps
Type "clear"
Enter
Sleep 300ms
Type "./bin/faultline fix examples/runtime-mismatch.log"
Enter
Wait+Screen /Fix Steps/
Sleep 1800ms
Screenshot docs/readme-assets/demo-fix.png
Sleep 1600ms

# Step 3 — classify an entire failing pipeline in one pass
Type "clear"
Enter
Sleep 300ms
Type "./bin/faultline batch examples/missing-executable.log examples/runtime-mismatch.log examples/docker-auth.log examples/lockfile-drift.log"
Enter
Wait+Screen /4 distinct/
Sleep 2000ms
Screenshot docs/readme-assets/demo-batch.png
Sleep 2800ms
