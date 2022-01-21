package v2utils

import (
    "log"
    "regexp"
    "os"
    "math/rand"
    "time"
    "fmt"
)

// doing <Loglevel>x to avoid clashing with methods
type Logger struct {
    Logfh   *os.File
    Debugx  *log.Logger
    Infox   *log.Logger
    Warnx   *log.Logger
    Critx   *log.Logger
}

type LogBannerModifier func(l *Logger)
type LogBannerType string

type LogDebugEntry func(string)
type LogInfoEntry func(string)
type LogWarnEntry func(string)
type LogCritEntry func(string)

var Debug LogBannerType = "debug"
var Info  LogBannerType = "info"
var Warn  LogBannerType = "warn"
var Crit  LogBannerType = "crit"

func NewLogger(logfile string, banners ...LogBannerModifier) *Logger {
    fh, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    flags := log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile

    l := &Logger{
        Logfh:  fh,
        Debugx: log.New(fh, "DEBUG: ", flags),
        Infox:  log.New(fh, "INFO: ", flags),
        Warnx:  log.New(fh, "WARN: ", flags),
        Critx:  log.New(fh, "CRIT: ", flags),
    }

    for _, mod := range banners {
        mod(l)
    }

    return l
}

func (l *Logger) Close() {
    defer l.Logfh.Close()

    l.Debugx = nil
    l.Infox = nil
    l.Warnx = nil
    l.Critx = nil
}

func (l *Logger) GetLogUtilities() (LogDebugEntry, LogInfoEntry, LogWarnEntry, LogCritEntry) {
    return l.getDebug(), l.getInfo(), l.getWarn(), l.getCrit()
}

func (l *Logger) getDebug() LogDebugEntry {
    return func(s string) { l.Debugx.Print(s) }
}

func (l *Logger) getInfo() LogInfoEntry {
    return func(s string) { l.Infox.Print(s) }
}

func (l *Logger) getWarn() LogWarnEntry {
    return func(s string) { l.Warnx.Print(s) }
}

func (l *Logger) getCrit() LogCritEntry {
    return func(s string) { l.Critx.Print(s) }
}


//
// Log Banner Updaters + Setters

func (l *Logger) UpdateDebugBanner(s string) {
    l.UpdateBanner(Debug, s)
}

func (l *Logger) UpdateInfoBanner(s string) {
    l.UpdateBanner(Info, s)
}

func (l *Logger) UpdateWarnBanner(s string) {
    l.UpdateBanner(Warn, s)
}

func (l *Logger) UpdateCritBanner(s string) {
    l.UpdateBanner(Crit, s)
}

func (l *Logger) UpdateBanner(t LogBannerType, s string) {
    mod := ModBanner(t, s)
    mod(l)
}

func SetDebugBanner(s string) LogBannerModifier {
    return ModBanner(Debug, s)
}

func SetInfoBanner(s string) LogBannerModifier {
    return ModBanner(Info, s)
}

func SetWarnBanner(s string) LogBannerModifier {
    return ModBanner(Warn, s)
}

func SetCritBanner(s string) LogBannerModifier {
    return ModBanner(Crit, s)
}

func ModBanner(t LogBannerType, s string) LogBannerModifier {
    empty := regexp.MustCompile(`^\s*$`)
    if empty.MatchString(s) {
        log.Fatalf("Invalid/empty banner('%s')", s)
    }

    // remove ending spaces and/or add single space at the end
    ending := regexp.MustCompile(`[\s\t\n]*$`)
    s = ending.ReplaceAllString(s, " ")

    var ret LogBannerModifier 
    switch t {
        case Debug:
            ret = func(l *Logger) {
                l.Debugx.SetPrefix(s)
            }
        case Info:
            ret = func(l *Logger) {
                l.Infox.SetPrefix(s)
            }
        case Warn:
            ret = func(l *Logger) {
                l.Warnx.SetPrefix(s)
            }
        case Crit:
            ret = func(l *Logger) {
                l.Critx.SetPrefix(s)
            }
        default:
            log.Fatalf("Invalid banner type(%s)", t)
    } 

    return ret
}


//
// Prettiness

func GetPrettyLogID() string {
    // Do yourself a favour and add some of your
    // fav names and adjectives below..

    name := []string{
        "andrea", "andreas", "andres", "anatol", "anton", "azarel", "aura", "andromeda", "adam", "ashel", "arnold",
        "blanka", "bayaka", "brett", "bol", "ben", "billy", "bozo", "bonifac", "barrack", "bozena", "beatrix", "bob",
        "carlos", "cleopatra", "coco",
        "deanna", "dana", "dolly", "dinesha",
        "elvira", "emmanuel", "emma", "eleanor", "ernesto",
        "fiona", "frank", "fidel",
        "greg", "george", "gana", "gabina", "gillian", "gary", "gazza",
        "hera", "helena", "hanka",
        "ivan", "itta", "ivana", "idiot",
        "josef", "jessica", "junior", "june",
        "kate", "katarina", "keanu",
        "linda", "leona", "leo", "leonardo",
        "mickey", "milan", "maria", "minnie", "margita", "melinda",
        "nickolas", "nataly", "nathan",
        "oliver", "olivia", "omar",
        "paul", "paula", "paola", "patsy",
        "quido",
        "rowell", "riahnna", "romeo",
        "sundar", "swan", "stella", "serena", "simone", "sandy", "sienna", "sharon", "shazza",
        "travis", "tom", "tim", "tam", "tamara", "tina", "teresa",
        "uma", "ursula", "uta", "uzi",
        "vella", "vikram", "vince", "varel", "venus", "viktor", "vanessa", "victoria",
        "wanger", "wally", "william",
        "yolanda",
        "ziggy",
    }

    adjective := [] string{
        "awed", "angry", "abismal", "atypical", "annoying",
        "bonkers", "balmy", "brisk", "beautiful", "bitchy",
        "cold", "clever", "clear", "carnivorous", "chubby", "cheeky",
        "dull", "dense", "dirty", "dear", "deep", "dim", "dark", "dire", "dopey", "dodgy",
        "ecstatic", "enamored", "elated", "expensive", "eternal", "enticed",
        "funny", "flimsy", "factored", "fresh", "fishy", "fidgetty", "fast", "fussy", "filthy", "funky", "faisty", "fair",
        "gentle", "giving", "ghastly", "girly",
        "hot", "hard", "hilly", "humid", "horny", "hungry", "holy",
        "introverted", "indisposable", "inexpensive", "impotent", "important", "invisible",
        "jovial", "junior", "jealous",
        "keen", "ketonic",
        "lethargic", "lame", "long", "laddy", "luminous",
        "manic", "moist", "masterful", "mindful", "misty",
        "naked", "nilly", "nosy", "noisy",
        "old", "ominous", "omnipotent", "obscure",
        "patronizing", "precious", "pretentious", "polite",
        "quiet", "quantitative",
        "randy", "real", "rainy", "roasted",
        "silly", "smart", "sleepy", "skinny", "sloppy", "saint", "shifty", "stoned", "sticky", "stinky", "slow", "scary", "scared", "salty",
        "tense", "tangy", "tiny", "troublesome", "technical",
        "unimportant", "unholy",
        "vegetarian", "vegan", "valuable", "vain", "visible", "violent",
        "warm", "wholesome", "witty", "warring", "watery", "weary", "windy",
        "young", "yellow",
        "zigzaggy", "zoned",
    }

    rand.Seed(time.Now().UnixNano())
    nindex := rand.Intn(len(name)-1)
    aindex := rand.Intn(len(adjective)-1)

    return fmt.Sprintf("%s_%s", adjective[aindex], name[nindex])
}
