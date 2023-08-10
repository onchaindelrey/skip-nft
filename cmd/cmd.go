package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	collection "skip-nft/collection"
	"sync"
)

const (
	COLOR_GREEN = "\033[32m"

	COLOR_BABY_BLUE = "\033[94m"

	COLOR_RED = "\033[31m"

	COLOR_YELLOW = "\033[33m"

	COLOR_RESET = "\033[0m"

	collectionNftCount = "collection-nft-count"
	collectionName     = "collection-name"
	maxProcessThreads  = "max-process-threads"
	baseUrl            = "base-url"
)

var (
	collectionNameDefault = "azuki1"
	collectionNftCountVal = 10000
	maxProcessThreadsVal  = 50
	baseUrlValueDefault   = "https://go-challenge.skip.money"

	Logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	rootCmd = &cobra.Command{
		Use:   "skip-collection-processor",
		Short: "Load and process rarity of NFT's from a provided collection",
		Run: func(cmd *cobra.Command, args []string) {
			err := runCollectionProcessor(cmd, args)
			if err != nil {
				Logger.Println(string(COLOR_RED), "ERROR:", err, string(COLOR_RESET))
			}
			Logger.Println(string(COLOR_YELLOW), "😎👍🕊️ \n", string(COLOR_RESET))
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&collectionNameDefault, collectionName, "n", "azuki1", "Name of the collection")
	rootCmd.PersistentFlags().StringVarP(&baseUrlValueDefault, baseUrl, "u", "https://go-challenge.skip.money", "Source url of the NFT collection")
	rootCmd.PersistentFlags().IntVarP(&collectionNftCountVal, collectionNftCount, "c", 10000, "Number of NFT's in the given collections")
	rootCmd.PersistentFlags().IntVarP(&maxProcessThreadsVal, maxProcessThreads, "t", 50, "Maximum number of parallel threads")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCollectionProcessor(cmd *cobra.Command, args []string) error {

	name, err := cmd.Flags().GetString(collectionName)
	if err != nil {
		return err
	}

	baseColUrl, err := cmd.Flags().GetString(baseUrl)
	if err != nil {
		return err
	}

	collectionNftCount, err := cmd.Flags().GetInt(collectionNftCount)
	if err != nil {
		return err
	}

	var processThreads = 50

	maxThreads, err := cmd.Flags().GetInt(maxProcessThreads)
	if err != nil {
		Logger.Println(string(COLOR_YELLOW), "WARNING: maximum number of process threads not set; running default of 10", string(COLOR_RESET))
	}

	if maxThreads > 0 {
		processThreads = maxThreads
	}

	azuki := collection.Collection{
		Count:             collectionNftCount,
		Name:              name,
		BaseUrl:           baseColUrl,
		Tokens:            make([]*collection.Token, collectionNftCount),
		TraitsList:        make(map[string][]string),
		TokenRarityScores: make([]collection.RarityScorecard, collectionNftCount),
		Mutex:             sync.RWMutex{},
	}

	azuki.LoadTokenCollection(processThreads)

	azuki.CalculateCollectionTokenRarity(processThreads)

	topFive := azuki.GetTopFive()

	topFiveJson, err := json.MarshalIndent(topFive, "", "    ")

	if err != nil {
		return err
	}

	Logger.Println(string(COLOR_BABY_BLUE), fmt.Sprintf("\n Top 5 rarest NFTs from collection : \n %s", string(topFiveJson)), string(COLOR_RESET))

	return nil
}
