package blockchain

import "github.com/deroproject/derohe/config"

func CalcSupply(height uint64) uint64 {
	supply := config.PREMINE
	remainingBlocks := height
	epochStartHeight := uint64(0)

	for remainingBlocks > 0 {
		blocksInEpoch := RewardReductionInterval
		if remainingBlocks < blocksInEpoch {
			blocksInEpoch = remainingBlocks
		}

		supply += CalcBlockReward(epochStartHeight) * blocksInEpoch

		remainingBlocks -= blocksInEpoch
		epochStartHeight += RewardReductionInterval
	}

	return supply
}
