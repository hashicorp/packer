There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

-   `<bs>` - Backspace

-   `<del>` - Delete

-   `<enter>` and `<return>` - Simulates an actual "enter" or "return" keypress.

-   `<esc>` - Simulates pressing the escape key.

-   `<tab>` - Simulates pressing the tab key.

-   `<f1>` - `<f12>` - Simulates pressing a function key.

-   `<up>` `<down>` `<left>` `<right>` - Simulates pressing an arrow key.

-   `<spacebar>` - Simulates pressing the spacebar.

-   `<insert>` - Simulates pressing the insert key.

-   `<home>` `<end>` - Simulates pressing the home and end keys.

-   `<pageUp>` `<pageDown>` - Simulates pressing the page up and page down keys.

-   `<leftAlt>` `<rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl>` `<rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift>` `<rightShift>` - Simulates pressing the shift key.

-   `<leftAltOn>` `<rightAltOn>` - Simulates pressing and holding the alt key.

-   `<leftCtrlOn>` `<rightCtrlOn>` - Simulates pressing and holding the ctrl key.

-   `<leftShiftOn>` `<rightShiftOn>` - Simulates pressing and holding the shift key.

-   `<leftAltOff>` `<rightAltOff>` - Simulates releasing a held alt key.

-   `<leftCtrlOff>` `<rightCtrlOff>` - Simulates releasing a held ctrl key.

-   `<leftShiftOff>` `<rightShiftOff>` - Simulates releasing a held shift key.

-   `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

-   `<waitXX>` - Add user defined time.Duration pause before sending any
    additional keys. For example `<wait10m>` or `<wait1m20s>`

When using modifier keys `ctrl`, `alt`, `shift` ensure that you release them,
otherwise they will be held down until the machine reboots. Use lowercase
characters as well inside modifiers.

For example: to simulate ctrl+c use `<leftCtrlOn>c<leftCtrlOff>`.

In addition to the special keys, each command to type is treated as a
[template engine](/docs/templates/engine.html). The
available variables are:

-   `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank!

For more examples of various boot commands, see the sample projects from our
[community templates page](/community-tools.html#templates).

