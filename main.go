package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cgt.name/pkg/go-mwclient"
	"git.omegaice.dev/atitd/research-bot/chatlog"
	"github.com/bwmarrin/discordgo"
)

var HomeGuild *discordgo.Guild = nil

var Session, _ = discordgo.New()
var Wiki, _ = mwclient.New("https://atitd.wiki/tale10/w/api.php", "OmegaiceGenerator")
var Blacklist []string

var username = ""
var password = ""

func init() {
	Session.Token = os.Getenv("DG_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Authentication Token")
	}

	username = os.Getenv("WIKI_USERNAME")
	if username == "" {
		flag.StringVar(&username, "username", "", "")
	}

	password = os.Getenv("WIKI_PASSWORD")
	if password == "" {
		flag.StringVar(&password, "password", "", "")
	}

	if err := Wiki.Login(username, password); err != nil {
		log.Println(err)
	}

	// if err := SetBot(Wiki, true); err != nil {
	// 	log.Fatalln(err)
	// }
}

func SetBot(w *mwclient.Client, value bool) error {
	token, err := w.GetToken("userrights")
	if err != nil {
		log.Fatalln(err)
	}

	cParameters := map[string]string{
		"action": "userrights",
		"user":   "Omegaice",
		"reason": "Bulk updates",
		"token":  token,
	}

	if value {
		cParameters["add"] = "bot"
	} else {
		cParameters["remove"] = "bot"
	}

	cResp, err := w.Post(cParameters)
	if err != nil {
		return err
	}
	log.Println(cResp)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	if Session.Token == "" {
		log.Fatalln("You must provide a Discord authentication token.")
	}

	if err := Session.Open(); err != nil {
		log.Fatalln("error opening connection to Discord", err)
	}
	defer Session.Close()

	log.Println("Current Guilds:")
	for _, guild := range Session.State.Guilds {
		log.Println("\t", guild.ID, guild.Name)
	}

	Session.AddHandler(onBlocked)
	Session.AddHandler(onUnblocked)
	Session.AddHandler(onComplete)

	api := chatlog.New()
	api.AddHandler("system", onTechStart)
	api.AddHandler("system", onTechComplete)
	api.AddHandler("system", onLabConstructed)
	go api.Tail()

	// Wait to exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	Session.Close()
}
