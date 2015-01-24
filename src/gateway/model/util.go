package model

func popID(queryID int64, searchIDs []int64) (ids []int64, found bool) {
	found = false
	for _, id := range searchIDs {
		if id == queryID {
			found = true
		} else {
			ids = append(ids, id)
		}
	}
	return
}
