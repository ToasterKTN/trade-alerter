package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/hpcloud/tail"
)

func main() {
	logfile := flag.String("LogFile", "C:/Program Files (x86)/Steam/steamapps/common/Path of Exile/logs/Client.txt", "Logfile to parse for")
	beep := flag.Bool("Beep", false, "Beep instead of play Sound")
	soundFile := flag.String("SoundFile", "sound.mp3", "MP3 File to play as sound if Beep is false")
	flag.Parse()
	lastbeep := time.Now().Unix()
	c := tail.Config{Follow: true, ReOpen: true, MustExist: true, Poll: true}
	c.Location = &tail.SeekInfo{Offset: -1, Whence: io.SeekEnd}
	t, err := tail.TailFile(*logfile, c)
	if err != nil {
		panic(err.Error())
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
		if strings.Contains(line.Text, "@From") {
			if lastbeep < time.Now().Unix()-10 {
				doBeep(*beep, *soundFile)
				lastbeep = time.Now().Unix()
			}
		}
	}
}

var (
	beepFunc = syscall.MustLoadDLL("user32.dll").MustFindProc("MessageBeep")
)

func doBeep(beep bool, sound string) {
	if beep {
		beepFunc.Call(0xffffffff)
	} else {
		playSound(sound)
	}
}

func playSound(sound string) {
	f, err := os.Open(sound)
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
