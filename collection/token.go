package collection

type Token struct {
	ID    int
	Attrs map[string]string
}

func (t *Token) GetTokenTraits() map[string]string {
	traitsList := make(map[string]string)

	for key, val := range t.Attrs {
		traitsList[key] = val
	}
	return traitsList
}

func (t *Token) CalculateTokenRarity(collectionTokens []*Token, collectionTraitsList map[string][]string) RarityScorecard {

	tokenRarity := float64(0)

	for traitName, traitValue := range t.Attrs {

		traitCategoryValuesCount := float64(len(collectionTraitsList[traitName]))

		var otherTokensWithValue float64 = 1

		for _, collectionToken := range collectionTokens {

			if t.ID == collectionToken.ID {
				continue
			}
			if _, exists := collectionToken.Attrs[traitName]; exists {

				if collectionToken.Attrs[traitName] == traitValue {
					otherTokensWithValue++
				}

			}

		}

		tokenRarity += float64(1) / (otherTokensWithValue * traitCategoryValuesCount)

	}

	return RarityScorecard{
		ID:     t.ID,
		Rarity: tokenRarity,
	}
}
