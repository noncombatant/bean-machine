# Bean Machine

Bean Machine is a web application that plays audio and video in a browser from
files served from a server. The server can be on a local intranet or on the
Internet. This software package contains both components: the server software,
and the HTML, CSS, and JavaScript that run in the browser.

## How To Install And Run Bean Machine

### 1: Install The Go Language

The Bean Machine server component is written in a programming language called
Go. Go does not (yet) come installed by default as part of your computer’s
operating system, but you can install it for free.

To run Bean Machine, you will need to [**download the Go programming
language**](https://golang.org/dl/) for your computer. Download the **latest
stable version** (at least version 1.14) of the **installer package**, and
install it.

### 2: Build The Server

Then you’ll need to build the server executable for your machine.

**First**, open a command shell terminal. (That’s Terminal on macOS, or CMD.EXE
on Windows.)

**Second**, change the shell’s working directory to the location of this
README.md file. This example assumes you’ve downloaded it into your Downloads
folder, which you probably have.

#### macOS And Linux

```
cd ~/Downloads/bean-machine
go build
```

#### Windows

```
cd %HOMEPATH%\Downloads\bean-machine
go build
```

The result of this process will be a file named bean-machine (macOS and Linux)
or bean-machine.exe (Windows). This is the Bean Machine server program. Its
purpose is to scan your music collection to build a catalog, and then to serve
that catalog and your music to web browsers. The actual music-playing user
interface appears in the browser.

### 3. Run The Server

Run bean-machine, to tell it to build the music catalog and run the web server.
Tell it the pathname of your music directory. On macOS, many Linux systems, and
Windows, that is most likely to be your Music folder.

#### macOS And Linux

```
./bean-machine -m ~/Music serve
```

#### Windows

```
.\bean-machine.exe -m %HOMEPATH\Music serve
```

bean-machine will scan your music directory (this might take a while) and then
print out the URL(s) by which you can access it.

#### Important Note

Finally, note that when you browse to your Bean Machine server, your browser
will warn you about the ‘invalid’ server security certificate that Bean Machine
uses. Bean Machine creates a new certificate for itself as part of starting the
server, but the certificate has not been *signed* by an authority your browser
knows about. Normally, you shouldn’t expect to see invalid certificates on real,
public internet sites.

But for running servers (like Bean Machine) on your own computer at home, it’s
OK, and you can click through this warning without hurting anything. Although
you know you are talking to your own server, the browser doesn’t ‘know’ that,
and so it warns you out of an abundance of caution.

(For public web sites like facebook.com or google.com, such a warning would be
important and real, and you should not click through it!)

## Help

In the main page of the application, you can show a help screen by typing **?**
or **h**.

## TODO

Consider using https://pkg.go.dev/embed.
