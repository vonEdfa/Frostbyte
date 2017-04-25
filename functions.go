package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"strings"

	"github.com/bwmarrin/discordgo"
	input "github.com/vonEdfa/go-input"
)

// Random - Picks a random integer.
// min: Minimum amount in the integer
// max: Maximum amount in the integer
func Random(min, max int) int {
	switch min {
	case max:
		return min
	default:
		rand.Seed(time.Now().Unix())
		return rand.Intn(max - min)
	}
}

// CheckToken - Check to see if the given token exists.
// token: The token of the bot we're trying to connect to
func CheckToken(token string) error {
	test, _ := discordgo.New("Bot " + token)
	if _, err := test.User("@me"); err != nil {
		test.Close()
		return fmt.Errorf("Failed to find Bot. Is the token correct?")
	}
	err := test.Close()
	if err != nil {
		veeLog(LogWarning, "Failed to close test connection: %v", err)
	}
	return nil
}

// GetPageContents - Get page content based on URL.
// url: Valid url of image.
func GetPageContents(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// IsManager Check to see if a user has ManageServer Permissions.
// s: The Current Session between the bot and discord
// m: The Message Object sent back from Discord.
func IsManager(s *discordgo.Session, GuildID string, AuthorID string) bool {
	// Check the user permissions of the guild.
	perms, err := s.State.UserChannelPermissions(AuthorID, GuildID)
	if err == nil {
		if (perms & discordgo.PermissionManageServer) > 0 {
			return true
		}
	} else {
		return false
	}
	return false
}

func (bot *Object) Init() {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	fmt.Println(`---------------------------------------------`)
	fmt.Printf("INSTALLING BOT\n")
	fmt.Println(`---------------------------------------------`)
	fmt.Printf("\nPlease fill in the below settings in order to complete setup.\nHINT: Go to https://discordapp.com/developers/applications/me to find your bot token.\n\n")

	// Ask for Bot Token.
	token, err := ui.Ask("Bot Token", &input.Options{
		Required:    true,
		Prefix:      "\t",
		ErrSuffix:   "\n",
		Loop:        true,
		LoopOrder:   "Try Again",
		HideOrder:   true,
		HideDefault: true,
		ValidateFunc: func(s string) error {
			token := strings.Replace(s, "\r", "", -1)
			if err := CheckToken(token); err != nil {
				return fmt.Errorf("* %s", err)
			}
			return nil
		},
		HideValidateFuncErr: true,
	})
	if err != nil {
		veeLog(LogWarning, "Input: %v", err)
		veeLog(LogInformational, "Config aborted. Shutting down.")
		//os.Exit(0)
	}

	// Ask for Guild ID.
	guild1, err := ui.Ask("Guild/Server ID", &input.Options{
		Required:    true,
		Prefix:      "\t",
		ErrSuffix:   "\n",
		Loop:        true,
		LoopOrder:   "Try Again",
		HideOrder:   true,
		HideDefault: true,
	})
	if err != nil {
		veeLog(LogWarning, "Input: %v", err)
		veeLog(LogInformational, "Config aborted. Shutting down.")
		//os.Exit(0)
	}

	// Cleanup and set the token and guild ID we've recieved.
	bot.Token = strings.Replace(token, "\r", "", -1)
	bot.Guild = strings.Replace(guild1, "\r", "", -1)

	// Save config.
	if err == nil {
		conf, err := json.MarshalIndent(bot, "", "  ")
		if err != nil {
			veeLog(LogError, "%v", err)
		} else {
			ioutil.WriteFile("config.json", conf, 0777)
		}
	}
}

// Save - Saves Database to config.json
// bot: Main Object with all your settings.
// s: The Current Session between the bot and discord
// m: The Message Object sent back from Discord.
func (bot *Object) Save() {
	for {
		<-time.After(5 * time.Minute)
		js, err := json.MarshalIndent(bot, "", "  ")
		if err == nil {
			ioutil.WriteFile("config.json", js, 0777)
		}
	}
}

/* Not functional yet
func (bot *Object) PruneMessages() {
	for {
		<-time.After(1 * time.Hour)
		for _, m := range bot.System.Messages {
			if m.Timestamp < time.Now()-(3600*24*7) {

			}
		}
	}
}
*/

// GetRoleID - Grabs the Discord Role ID
// bot: Main Object with all your settings.
// s: The Current Session between the bot and discord
// role: The Discord role
func (bot *Object) GetRoleID(s *discordgo.Session, role string) string {
	var id string
	r, err := s.State.Guild(bot.Guild)
	if err == nil {
		for _, v := range r.Roles {
			if v.Name == role {
				id = v.ID
			}
		}
	}
	return id
}

// MemberHasRole - Checks to see if the user has a role.
// bot: Main Object with all your settings.
// s: The Current Session between the bot and discord
// role: The Discord role
func (bot *Object) MemberHasRole(s *discordgo.Session, AuthorID string, role string) bool {
	therole := bot.GetRoleID(s, role)
	z, err := s.State.Member(bot.Guild, AuthorID)
	if err != nil {
		z, err = s.GuildMember(bot.Guild, AuthorID)
		if err != nil {
			fmt.Println("Error ->", err)
			return false
		}
	}
	for r := range z.Roles {
		if therole == z.Roles[r] {
			return true
		}
	}
	return false
}

// Register - Register new object.
// bot: Main Object with all your settings.
// s: The Current Session between the bot and discord
// m: Message Object sent back from Discord.
func (bot *Object) Register(s *discordgo.Session, m *discordgo.MessageCreate) {
	// check and make sure the server already exists in my collection.
	if bot.System != nil {
		return
	}
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Guild = c.GuildID
	chn := &Channels{
		Autorole: "",
		Greeting: "",
		ByeMsg:   "",
	}

	// Create a new Info pointer.
	info := &System{
		Prefix:   ".",
		Autorole: "",
		Greeting: "",
		ByeMsg:   "",
		Channels: chn,
	}
	// Add our Info object to the bot map.
	bot.System = info
}

// Task - Store new messages to object.
// bot: Main Object with all your settings.
// s: The Current Session between the bot and discord
// role: The Discord role
func (bot *Object) Task(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Don't track the bots messages.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if bot.System == nil {
		return
	}
	// Create a new pointer.
	msg := &Messages{
		ID:        m.ID,
		Author:    m.Author.ID,
		Channel:   m.ChannelID,
		Timestamp: time.Now().Unix(),
	}
	// Add this Message to our Info object.
	bot.System.Messages = append(bot.System.Messages, msg)
}

// AddStatus - Adds a status string to the main Object.
// message: the status message(s)
func (bot *Object) AddStatus(message string) error {
	if bot.System == nil {
		return nil
	}
	for _, s := range bot.System.Status {
		if s == message {
			return errors.New("status exists already")
		}
	}
	bot.System.Status = append(bot.System.Status, message)
	return nil
}

// RemoveStatus - Removes a status string from the main Object.
// message: the status message
func (bot *Object) RemoveStatus(message string) error {
	if bot.System == nil {
		return errors.New("object doesn't exist")
	}
	var ti int
	for i, k := range bot.System.Status {
		if k == message {
			ti = i
		}
	}
	if ti == 0 {
		return errors.New("status doesn't exist in my collection")
	}
	bot.System.Status = append(bot.System.Status[:ti], bot.System.Status[ti+1:]...)
	return nil
}

// Banner - Displays the bot banner upon startup.
func (bot *Object) Banner() {
	fmt.Println(`
 Y88b      /                     
  Y88b    / e88~~8e   e88~~8e  
   Y88b  / d888  88b d888  88b 
    Y888/  8888__888 8888__888 
     Y8/   Y888    , Y888    , 
      Y     "88___/   "88___/ `)
	fmt.Printf("\n Powered by Frostbyte and DiscordGo!\n Version: %v\n\n", VERSION)
	<-time.After(2 * time.Second)
}

// Intro - Displays introduction and information on startup.
func (bot *Object) Intro(s *discordgo.Session) {
	var ars map[string]string
	fmt.Println(`---------------------------------------------`)
	// Collect some information and display it!
	guild, err := s.State.Guild(bot.Guild)
	if err != nil {
		fmt.Println(err)
	} else {
		// Channel count for server
		fmt.Print(len(guild.Channels), " Channel(s), ")
		// Member count for server
		fmt.Print(len(guild.Members), " Member(s), ")
		// Role count for server
		fmt.Print(len(guild.Roles), " Role(s), ")
	}

	// Collect the A.R.S Count.
	io, err := ioutil.ReadFile("autoresponse.json")
	if err != nil {
		fmt.Println(err)
	} else {
		json.Unmarshal(io, &ars)
		// ARS Count for the bot
		fmt.Println(len(ars), " A.R.S Rule(s).")
	}
	fmt.Println(`---------------------------------------------`)
}
