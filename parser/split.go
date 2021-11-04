package parser

/*
	from https://github.com/fatih/camelcase
	The MIT License (MIT)
	Copyright (c) 2015 Fatih Arslan
*/

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//   "" =>                     [""]
//   "lowercase" =>            ["lowercase"]
//   "Class" =>                ["Class"]
//   "MyClass" =>              ["My", "Class"]
//   "MyC" =>                  ["My", "C"]
//   "HTML" =>                 ["HTML"]
//   "PDFLoader" =>            ["PDF", "Loader"]
//   "AString" =>              ["A", "String"]
//   "SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//   "vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//   "GL11Version" =>          ["GL", "11", "Version"]
//   "99Bottles" =>            ["99", "Bottles"]
//   "May5" =>                 ["May", "5"]
//   "BFG9000" =>              ["BFG", "9000"]
//   "BöseÜberraschung" =>     ["Böse", "Überraschung"]
//   "Two  spaces" =>          ["Two", "  ", "spaces"]
//   "BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1) If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2) Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3) Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4) Iterate through array of split strings, and if a given string
//     is upper case:
//       if subsequent string is lower case:
//         move last character of upper case string to beginning of
//         lower case string
func Split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}

// camelizeDown converts a name or other string into a camel case
// version with the first letter lowercase. "ModelID" becomes "modelID".
func camelizeDown(word string) string {
	if isAcronym(word) {
		// entire word is an acronym
		return strings.ToLower(word)
	}
	words := Split(word)
	for i := range words {
		if isAcronym(words[i]) {
			if i == 0 {
				words[i] = strings.ToLower(words[i])
			} else {
				words[i] = strings.ToUpper(words[i])
			}
		}
	}
	word = strings.Join(words, "")
	return strings.ToLower(word[:1]) + word[1:]
}

// snakeDown converts a name or other string into a snake case
// version with all letters lowercase and use underscore for spaces.
// "ModelID" becomes "model_id".
func snakeDown(word string) string {
	if isAcronym(word) {
		// entire word is an acronym
		return strings.ToLower(word)
	}

	var res = make([]rune, 0, len(word))
	var p = '_'

	for i, r := range word {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '_')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '_', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}

	return string(res)
}

func isAcronym(word string) bool {
	for _, ac := range baseAcronyms {
		if strings.EqualFold(ac, word) {
			return true
		}
	}
	return false
}

var baseAcronyms = strings.Split(`HTML,JSON,JWT,ID,UUID,SQL,ACK,ACL,ADSL,AES,ANSI,API,ARP,ATM,BGP,BSS,CAT,CCITT,CHAP,CIDR,CIR,CLI,CPE,CPU,CRC,CRT,CSMA,CMOS,DCE,DEC,DES,DHCP,DNS,DRAM,DSL,DSLAM,DTE,DMI,EHA,EIA,EIGRP,EOF,ESS,FCC,FCS,FDDI,FTP,GBIC,gbps,GEPOF,HDLC,HTTP,HTTPS,IANA,ICMP,IDF,IDS,IEEE,IETF,IMAP,IP,IPS,ISDN,ISP,kbps,LACP,LAN,LAPB,LAPF,LLC,MAC,MAN,Mbps,MC,MDF,MIB,MoCA,MPLS,MTU,NAC,NAT,NBMA,NIC,NRZ,NRZI,NVRAM,OSI,OSPF,OUI,PAP,PAT,PC,PIM,PIM,PCM,PDU,POP3,POP,POTS,PPP,PPTP,PTT,PVST,RADIUS,RAM,RARP,RFC,RIP,RLL,ROM,RSTP,RTP,RCP,SDLC,SFD,SFP,SLARP,SLIP,SMTP,SNA,SNAP,SNMP,SOF,SRAM,SSH,SSID,STP,SYN,TDM,TFTP,TIA,TOFU,UDP,URL,URI,USB,UTP,VC,VLAN,VLSM,VPN,W3C,WAN,WEP,WiFi,WPA,WWW`, ",")
