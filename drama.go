package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
	"math/rand"
	"os"
	"strings"
)

type Words struct {
	Peoples    []string   `json:"peoples"`
	Categories [][]string `json:"categories"`
}

type Config struct {
	Token     string `json:"token"`
	OpenAiKey string `json:"openai_key"`
	GuildID   string `json:"guild_id"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config: ", err)
		return
	}

	words, err := loadWords()
	if err != nil {
		fmt.Println("Error loading words: ", err)
		return
	}

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	err = createCommand(discord, config.GuildID)
	if err != nil {
		fmt.Println("Error creating command: ", err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == "generate" {
			err := generateDescription(s, i, words, config)
			if err != nil {
				return
			}
		}
	})

	fmt.Printf("Bot is now running on %d servers", len(discord.State.Guilds))
	select {}
}

func loadConfig() (*Config, error) {
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return &config, nil
}

func loadWords() (*Words, error) {
	bytes, err := os.ReadFile("words.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var words Words
	err = json.Unmarshal(bytes, &words)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return &words, nil
}

func createCommand(discord *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "generate",
		Description: "Generuje opis dramatycznej sytuacji",
	}

	_, err := discord.ApplicationCommandCreate(discord.State.User.ID, guildID, command)
	if err != nil {
		return fmt.Errorf("error creating command: %v", err)
	}

	return nil
}

func generateDescription(s *discordgo.Session, i *discordgo.InteractionCreate, words *Words, config *Config) error {
	var wordsArr = [3]string{}

	for i := 0; i < 3; i++ {
		categoryIndex := rand.Intn(len(words.Categories))
		wordIndex := rand.Intn(len(words.Categories[categoryIndex]))
		wordsArr[i] = words.Categories[categoryIndex][wordIndex]
	}

	peopleIndex := rand.Intn(len(words.Peoples))

	prompt := fmt.Sprintf("Proszę wygeneruj dla mnie TYLKO JEDEN opis dramatycznej sytuacji korzystając z wyrazów: %s, pamiętaj aby uwzględnić osobę: %s. Napisz to w jednym krótkim zdaniu.", strings.Join(wordsArr[:], ", "), words.Peoples[peopleIndex])

	client := openai.NewClient(config.OpenAiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       "EternalCode.pl - DramaGPT",
		Description: resp.Choices[0].Message.Content,
		Color:       0x00ff00,
	}

	reply := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}

	respond := s.InteractionRespond(i.Interaction, reply)
	if respond != nil {
		fmt.Println(respond)
	}
	return nil
}
