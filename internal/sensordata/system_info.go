package sensordata

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"strings"
)

var screenSizes = [][]int{
	{1280, 720},
	{1920, 1080},
	{2560, 1440},
}

var androidVersionToSDK = map[string]int{
	"11":    30,
	"10":    29,
	"9":     28,
	"8.1.0": 27,
	"8.0.0": 26,
	"7.1":   25,
	"7.0":   24,
}

// SystemInfo represents Android system information
type SystemInfo struct {
	screenHeight          int
	screenWidth           int
	batteryCharging       bool
	batteryLevel          int
	orientation           int
	language              string
	androidVersion        string
	rotationLock          string
	buildModel            string
	buildBootloader       string
	buildHardware         string
	packageName           string
	androidID             string
	keyboard              int
	adbEnabled            bool
	buildVersionCodename  string
	buildVersionIncremental int
	buildVersionSDK       int
	buildManufacturer     string
	buildProduct          string
	buildTags             string
	buildType             string
	buildUser             string
	buildDisplay          string
	buildBoard            string
	buildBrand            string
	buildDevice           string
	buildFingerprint      string
	buildHost             string
	buildID               string
}

// NewSystemInfo creates a new SystemInfo
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{}
}

// Randomize generates random system information
func (s *SystemInfo) Randomize() {
	// Simplified - just use Pixel 3a with Android 11
	deviceModel := "Pixel 3a"
	codename := "sargo"
	buildVersion := "11"
	buildID := "RQ3A.210605.005"
	buildVersionIncremental := mathrand.Intn(9000000) + 1000000

	screenSize := screenSizes[mathrand.Intn(len(screenSizes))]
	s.screenHeight = screenSize[0]
	s.screenWidth = screenSize[1]
	s.batteryCharging = mathrand.Intn(10) <= 1
	s.batteryLevel = mathrand.Intn(80) + 10
	s.orientation = 1
	s.language = "en"
	s.androidVersion = buildVersion
	if mathrand.Intn(10) > 1 {
		s.rotationLock = "1"
	} else {
		s.rotationLock = "0"
	}
	s.buildModel = deviceModel
	s.buildBootloader = strconv.Itoa(mathrand.Intn(9000000) + 1000000)
	s.buildHardware = codename
	s.packageName = "com.interrait.mymazda"

	// Generate random Android ID
	androidIDBytes := make([]byte, 8)
	_, _ = rand.Read(androidIDBytes)
	s.androidID = hex.EncodeToString(androidIDBytes)

	s.keyboard = 0
	s.adbEnabled = false
	s.buildVersionCodename = "REL"
	s.buildVersionIncremental = buildVersionIncremental
	s.buildVersionSDK = androidVersionToSDK[buildVersion]
	s.buildManufacturer = "Google"
	s.buildProduct = codename
	s.buildTags = "release-keys"
	s.buildType = "user"
	s.buildUser = "android-build"
	s.buildDisplay = buildID
	s.buildBoard = codename
	s.buildBrand = "google"
	s.buildDevice = codename
	s.buildFingerprint = fmt.Sprintf("google/%s/%s:%s/%s/%d:user/release-keys",
		codename, codename, buildVersion, buildID, buildVersionIncremental)
	s.buildHost = fmt.Sprintf("abfarm-%d", mathrand.Intn(90000)+10000)
	s.buildID = buildID
}

// ToString converts SystemInfo to string format
func (s *SystemInfo) ToString() string {
	batteryChargingStr := "0"
	if s.batteryCharging {
		batteryChargingStr = "1"
	}

	adbEnabledStr := "0"
	if s.adbEnabled {
		adbEnabledStr = "1"
	}

	parts := []string{
		"-1",
		"uaend",
		"-1",
		strconv.Itoa(s.screenHeight),
		strconv.Itoa(s.screenWidth),
		batteryChargingStr,
		strconv.Itoa(s.batteryLevel),
		strconv.Itoa(s.orientation),
		percentEncode(s.language),
		percentEncode(s.androidVersion),
		s.rotationLock,
		percentEncode(s.buildModel),
		percentEncode(s.buildBootloader),
		percentEncode(s.buildHardware),
		"-1",
		s.packageName,
		"-1",
		"-1",
		s.androidID,
		"-1",
		strconv.Itoa(s.keyboard),
		adbEnabledStr,
		percentEncode(s.buildVersionCodename),
		percentEncode(strconv.Itoa(s.buildVersionIncremental)),
		strconv.Itoa(s.buildVersionSDK),
		percentEncode(s.buildManufacturer),
		percentEncode(s.buildProduct),
		percentEncode(s.buildTags),
		percentEncode(s.buildType),
		percentEncode(s.buildUser),
		percentEncode(s.buildDisplay),
		percentEncode(s.buildBoard),
		percentEncode(s.buildBrand),
		percentEncode(s.buildDevice),
		percentEncode(s.buildFingerprint),
		percentEncode(s.buildHost),
		percentEncode(s.buildID),
	}

	return strings.Join(parts, ",")
}

// GetCharCodeSum calculates the sum of character codes
func (s *SystemInfo) GetCharCodeSum() int {
	return sumCharCodes(s.ToString())
}

func percentEncode(str string) string {
	if str == "" {
		return ""
	}

	var sb strings.Builder
	for _, b := range []byte(str) {
		if b >= 33 && b <= 0x7E && b != 34 && b != 37 && b != 39 && b != 44 && b != 92 {
			sb.WriteByte(b)
		} else {
			sb.WriteString("%" + strings.ToUpper(fmt.Sprintf("%x", b)))
		}
	}
	return sb.String()
}

func sumCharCodes(str string) int {
	sum := 0
	for _, b := range []byte(str) {
		if b < 0x80 {
			sum += int(b)
		}
	}
	return sum
}
