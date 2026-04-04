Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 700
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 40ms

# -- Import fixture (binary pre-built) ------------------
Type "./skel import default-skel.json"
Enter
Sleep 2s
Ctrl+L

# -- list -----------------------------------------------
Type "./skel list"
Enter
Sleep 2s
Ctrl+L

# -- status ---------------------------------------------
Type "./skel status default"
Enter
Sleep 4s
Ctrl+L

# -- show -----------------------------------------------
Type "./skel show default"
Enter
Sleep 4s
Ctrl+L

# -- doctor ---------------------------------------------
Type "./skel doctor default"
Enter
Sleep 2s
Ctrl+L

# -- drift ----------------------------------------------
Type "./skel drift default"
Enter
Sleep 2s
Ctrl+L

# -- diff -----------------------------------------------
Type "./skel diff default default"
Enter
Sleep 2s
Ctrl+L

# -- restore --dry-run ----------------------------------
Type "./skel restore default --dry-run"
Enter
Sleep 2s
Ctrl+L

# -- export ---------------------------------------------
Type "./skel export default"
Enter
Sleep 2s
Ctrl+L

# -- brewfile export ------------------------------------
Type "./skel brewfile export default"
Enter
Sleep 2s
Ctrl+L

# -- update ---------------------------------------------
Type "./skel update default"
Enter
Sleep 2s
