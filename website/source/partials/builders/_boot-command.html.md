There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

-   `<bs>` - Backspace

-   `<del>` - Delete

-   `<enter> <return>` - Simulates an actual "enter" or "return" keypress.

-   `<esc>` - Simulates pressing the escape key.

-   `<tab>` - Simulates pressing the tab key.

-   `<f1> - <f12>` - Simulates pressing a function key.

-   `<up> <down> <left> <right>` - Simulates pressing an arrow key.

-   `<spacebar>` - Simulates pressing the spacebar.

-   `<insert>` - Simulates pressing the insert key.

-   `<home> <end>` - Simulates pressing the home and end keys.

-   `<pageUp> <pageDown>` - Simulates pressing the page up and page down keys.

-   `<leftAlt> <rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl> <rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift> <rightShift>` - Simulates pressing the shift key.

-   `<leftSuper> <rightSuper>` - Simulates pressing the ⌘ or Windows key.

-   `<wait> <wait5> <wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

-   `<waitXX>` - Add an arbitrary pause before sending any additional keys. The
    format of `XX` is a possibly signed sequence of decimal numbers, each with
    optional fraction and a unit suffix, such as `300ms`, `-1.5h` or `2h45m`.
    Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. For example
    `<wait10m>` or `<wait1m20s>`


### On/Off variants

Any printable keyboard character, and of these "special" expressions, with the
exception of the `<wait>` types, can also be toggled on or off. For example, to
simulate ctrl+c, use `<leftCtrlOn>c<leftCtrlOff>`. Be sure to release them,
otherwise they will be held down until the machine reboots.

To hold the `c` key down, you would use `<cOn>`. Likewise, `<cOff>` to release.

### Templates inside boot command

In addition to the special keys, each command to type is treated as a
[template engine](/docs/templates/engine.html). The
available variables are:

-   `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank!
*   `Name` - The name of the VM.

For more examples of various boot commands, see the sample projects from our
[community templates page](/community-tools.html#templates).

