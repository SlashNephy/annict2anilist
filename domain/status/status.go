package status

func IsSameListStatus(annict AnnictStatusState, aniList AniListMediaListStatus) bool {
	return aniList == annict.ToAniListStatus() || annict == aniList.ToAnnictStatus()
}
