package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	s				*discordgo.Session
	token =			os.Getenv("TOKEN_YANG")

	banYour =		[]string{
		"ur",
		"your",
		"you're",
		"u're",
		"u'r",
		"yo",
	}
	banMom =		[]string{
		"mom",
		"moom",
		"mooom",
		"moooom",
		"mooooom",

		"mother",
		"moother",
		"mooother",
		"moooother",
		"mooooother",

		"mommy",
		"moommy",
		"mooommy",
		"moooommy",
		"mooooommy",
		"mama",
		"mamma",
		"madre",
		"momma",
	}

	bads []bad
)



// A person who is being monitored as because they have been
// determined to potentially say "ur mom" in two messages in a row
type bad struct {
	guildID string    // Guild the message came from
	userID string
}

func init() {
	// Print boot "splash"
	fmt.Println("YangBot 1.0.0")
	log.Println("[Info] Minimum permissions are 1099780129792")
}

func main() {
	// Declare error here so it can be set without :=
	var err error
	
	// Create bot client session
	log.Println("Logging in")
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Could not create session: %v", err)
	}
	
	// Pass on message event to functions
	s.AddHandler(single)
	s.AddHandler(multiYour)
	s.AddHandler(multiMom)

	// We only care about message + guild member intents
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentMessageContent | discordgo.IntentGuildMessages | discordgo.IntentGuildMembers)

	// Open websocket connection to Discord and listen
	err = s.Open()
	if err != nil {
		log.Fatalf("Error opening websocket connection: %v", err)
	}

	// Close Discord session cleanly
	defer s.Close()
	
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stop
	
	log.Println("Ciao")
}

// Single message check
func single(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Iterate throughout every permutation of "your" and "mom" specified previously
	for _, your := range banYour {
		for _, mom := range banMom {
			// If message contains any specific permutation, timeout user
			if strings.Contains(strings.ToLower(m.Content), fmt.Sprintf("%s %s", your, mom)) {
				err := timeout(m.GuildID, m.Author.ID)
				if err != nil {
					return
				}
			}
		}
	}
}

// Implements the multi-message checking by using a global variable to keep
// track of a User ID which could possibly complete the trigger in two messages
// PART ONE - YOUR
func multiYour(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if already in list and ignore if yes
	if isBad(m.GuildID, m.Author.ID) {
		return
	}

	// If message matches first part, add to watchlist o.o
	for _, your := range banYour {
		if strings.ToLower(m.Content) == your {
			fresh := bad{m.GuildID, m.Author.ID}
			bads = append(bads, fresh)
		}
	}
}

func multiMom(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	if isBad(m.GuildID, m.Author.ID) {
		for _, mom := range banMom {
			if strings.ToLower(m.Content) == mom {
				timeout(m.GuildID, m.Author.ID)
			}
		}

		index, err := indexBad(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}

		removeBad(index)
	}
}

// Check if given member is a badUser
func isBad(g string, u string) (res bool) {
	for _, bad := range bads {
		if bad.guildID == g && bad.userID == u {
			return true
		}
	}

	return false
}

func indexBad(g string, u string) (index int, err error) {
	for i, bad := range bads {
		if bad.guildID == g && bad.userID == u {
			return i, nil
		} else {
			return -1, errors.New("bad user does not exist in list")
		}
	}

	return 0, errors.New("no bad users")
}

func removeBad(i int) {
	bads[i] = bads[len(bads) - 1]
	bads = bads[:len(bads) - 1]
}

func timeout(g string, u string) (err error) {
	// Get time five minutes in future
	until := time.Now().Add(time.Minute)

	err = s.GuildMemberTimeout(g, u, &until)
	if err != nil {
		return err
	}

	return nil
}
