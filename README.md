# todocli

A terminal based todo app for devs. Written in go using bubbletea, lipgloss and bubbles from charm\_.
Initially built this for myself because I used to write todos in a TODO.txt, which was inefficient as I'm mostly using my terminal.

# Installation

## Windows Guide:

Head over to the release page and download the todo.exe binary.

> [!WARNING]
> Windows may flag the binary as a virus, but that is a false positive, you can turn off defender if it causes issues while installing.

1. Move todo.exe to a desired folder on your computer
2. Search "Edit the system environment variables" and press enter.
3. Click "Environment Variables" and then double click PATH under the user heading.
4. Copy the path to your folder where you stored todo.exe
5. Add a new path and paste in the copied path.
6. Enter and keep clicking OK till you exit out of System Properties
7. You can now run the app by running the command "todo" from a terminal

## Linux Guide:

After downloading the binary from the release. (Please use the linux release)

1. Open the terminal in the place where you have stored todo binary i.e Downloads.
2. Run the following command:

```
sudo mv todo /usr/local/bin
```
