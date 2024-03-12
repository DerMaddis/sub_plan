package parser

import "strings"

type RoomCancelledData struct {
}

type RoomSwitchData struct {
	Room1 string
	Room2 string
}

type RoomOtherData struct {
	Data string
}

func parseRoom(s string) interface{} {
	switch {
	case strings.HasSuffix(s, "---"): // this could be changed to detect <s></s> tags around the room (or both)
		return RoomCancelledData{}
	case strings.Contains(s, "→"):
		rooms := strings.Split(s, "→")
		if len(rooms) != 2 {
			return RoomOtherData{s}
		}
		return RoomSwitchData{rooms[0], rooms[1]}
	default:
		return RoomOtherData{s}
	}
}
