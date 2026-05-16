// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import "encoding/json"

type (
	ShowServerInfo struct{}

	ShowServerInfoResponse struct {
		MaxPlayers            int64  `json:"MaxPlayers"`
		GameMode              string `json:"GameMode_s"`
		MapName               string `json:"MapName_s"`
		ServerName            string `json:"ServerName_s"`
		CoopServer            bool   `json:"COOPSERVER_b"`
		SearchKeywords        string `json:"SEARCHKEYWORDS_s"`
		GameVersion           string `json:"GameVersion_s"`
		Licensed              bool   `json:"LICENSEDSERVER_b"`
		Playtime              int64  `json:"PLAYTIME_I"`
		Flags                 int64  `json:"Flags_I"`
		MatchHopper           string `json:"MATCHHOPPER_s"`
		MatchTimeout          int64  `json:"MatchTimeout_d"`
		SessionTemplateName   string `json:"SESSIONTEMPLATENAME_s"`
		Password              bool   `json:"Password_b"`
		PlayerCount           int64  `json:"PlayerCount_I"`
		TagLanguages          string `json:"TagLanguage_s"`
		Region                string `json:"Region_s"`
		CurrentModLoadedCount int64  `json:"CurrentModLoadedCount_I"`
		AllModsWhitelisted    bool   `json:"AllModsWhitelisted_b"`
		Team1                 string `json:"TeamOne_s"`
		Team2                 string `json:"TeamTwo_s"`
		SessionPassword       string `json:"SessionPassword_s"`
		PlayerReserveCount    int64  `json:"PlayerReserveCount_I"`
		PublicQueueLimit      int64  `json:"PublicQueueLimit_I"`
		PublicQueue           int64  `json:"PublicQueue_I"`
		ReservedQueue         int64  `json:"ReservedQueue_I"`
		BeaconPort            int64  `json:"BeaconPort_I"`
		NextLayer             string `json:"NextLayer_s"`
	}
)

func (c ShowServerInfo) ToBody() string {
	return "ShowServerInfo"
}

func (ShowServerInfo) ParseBody(body string) (ShowServerInfoResponse, error) {
	return ParseShowServerInfo(body)
}

func ParseShowServerInfo(body string) (ShowServerInfoResponse, error) {
	var out ShowServerInfoResponse
	err := json.Unmarshal([]byte(body), &out)
	return out, err
}
