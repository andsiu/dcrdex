// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package dex

var symbolBipIDs map[string]uint32

// BipSymbolID returns the asset ID associated with a given ticker symbol.
// While there are a number of duplicate ticker symbols in the BIP ID list
// (cpc, cmt, xrd, dst, one, ask, ...), those are disambiguated in the bipIDs
// map here, so must be referenced with their bracketed suffix.
func BipSymbolID(symbol string) (uint32, bool) {
	if symbolBipIDs == nil {
		symbolBipIDs = make(map[string]uint32)
		for idx, sym := range bipIDs {
			symbolBipIDs[sym] = idx
		}
	}

	idx, found := symbolBipIDs[symbol]
	return idx, found
}

// BipIDSymbol returns the BIP ID for a given symbol.
func BipIDSymbol(id uint32) string {
	return bipIDs[id]
}

var bipIDs = map[uint32]string{
	0:     "btc",
	1:     "testnet",
	2:     "ltc",
	3:     "doge",
	4:     "rdd",
	5:     "dash",
	6:     "ppc",
	7:     "nmc",
	8:     "ftc",
	9:     "xcp",
	10:    "blk",
	11:    "nsr",
	12:    "nbt",
	13:    "mzc",
	14:    "via",
	15:    "xch",
	16:    "rby",
	17:    "grs",
	18:    "dgc",
	19:    "ccn",
	20:    "dgb",
	21:    "openassets",
	22:    "mona",
	23:    "clam",
	24:    "xpm",
	25:    "neos",
	26:    "jbs",
	27:    "zrc",
	28:    "vtc",
	29:    "nxt",
	30:    "burst",
	31:    "mue",
	32:    "zoom",
	33:    "vash",
	34:    "cdn",
	35:    "sdc",
	36:    "pkb",
	37:    "pnd",
	38:    "start",
	39:    "moin",
	40:    "exp",
	41:    "emc2",
	42:    "dcr",
	43:    "xem",
	44:    "part",
	45:    "arg",
	46:    "libertas",
	47:    "posw",
	48:    "shr",
	49:    "gcr",
	50:    "nvc",
	51:    "ac",
	52:    "btcd",
	53:    "dope",
	54:    "tpc",
	55:    "aib",
	56:    "edrc",
	57:    "sys",
	58:    "slr",
	59:    "smly",
	60:    "eth",
	61:    "etc",
	62:    "psb",
	63:    "ldcn",
	64:    "openchain",
	65:    "xbc",
	66:    "iop",
	67:    "nxs",
	68:    "insn",
	69:    "ok",
	70:    "brit",
	71:    "cmp",
	72:    "crw",
	73:    "bela",
	74:    "icx",
	75:    "fjc",
	76:    "mix",
	77:    "xvg",
	78:    "efl",
	79:    "club",
	80:    "richx",
	81:    "pot",
	82:    "qrk",
	83:    "trc",
	84:    "grc",
	85:    "aur",
	86:    "ixc",
	87:    "nlg",
	88:    "bitb",
	89:    "bta",
	90:    "xmy",
	91:    "bsd",
	92:    "uno",
	93:    "mtr",
	94:    "gb",
	95:    "shm",
	96:    "crx",
	97:    "biq",
	98:    "evo",
	99:    "sto",
	100:   "bigup",
	101:   "game",
	102:   "dlc",
	103:   "zyd",
	104:   "dbic",
	105:   "strat",
	106:   "sh",
	107:   "mars",
	108:   "ubq",
	109:   "ptc",
	110:   "nro",
	111:   "ark",
	112:   "usc",
	113:   "thc",
	114:   "linx",
	115:   "ecn",
	116:   "dnr",
	117:   "pink",
	118:   "atom",
	119:   "pivx",
	120:   "flash",
	121:   "zen",
	122:   "put",
	123:   "zny",
	124:   "unify",
	125:   "xst",
	126:   "brk",
	127:   "vc",
	128:   "xmr",
	129:   "vox",
	130:   "nav",
	131:   "fct",
	132:   "ec",
	133:   "zec",
	134:   "lsk",
	135:   "steem",
	136:   "xzc",
	137:   "rbtc",
	138:   "giftblock",
	139:   "rpt",
	140:   "lbc",
	141:   "kmd",
	142:   "bsq",
	143:   "ric",
	144:   "xrp",
	145:   "bch",
	146:   "nebl",
	147:   "zcl",
	148:   "xlm",
	149:   "nlc2",
	150:   "whl",
	151:   "erc",
	152:   "dmd",
	153:   "btm",
	154:   "bio",
	155:   "xwc",
	156:   "btg",
	157:   "btc2x",
	158:   "ssn",
	159:   "toa",
	160:   "btx",
	161:   "acc",
	162:   "bco",
	163:   "ella",
	164:   "pirl",
	165:   "xrb",
	166:   "vivo",
	167:   "frst",
	168:   "hnc",
	169:   "buzz",
	170:   "mbrs",
	171:   "hsr",
	172:   "html",
	173:   "odn",
	174:   "onx",
	175:   "rvn",
	176:   "gbx",
	177:   "btcz",
	178:   "poa",
	179:   "nyc",
	180:   "mxt",
	181:   "wc",
	182:   "mnx",
	183:   "btcp",
	184:   "music",
	185:   "bca",
	186:   "crave",
	187:   "stak",
	188:   "wbtc",
	189:   "lch",
	190:   "excl",
	191:   "lynx",
	192:   "lcc",
	193:   "xfe",
	194:   "eos",
	195:   "trx",
	196:   "kobo",
	197:   "hush",
	198:   "banano",
	199:   "etf",
	200:   "omni",
	201:   "bifi",
	202:   "ufo",
	203:   "cnmc",
	204:   "bcn",
	205:   "rin",
	206:   "atp",
	207:   "evt",
	208:   "atn",
	209:   "bis",
	210:   "neet",
	211:   "bopo",
	212:   "oot",
	213:   "xspec",
	214:   "monk",
	215:   "boxy",
	216:   "flo",
	217:   "mec",
	218:   "btdx",
	219:   "xax",
	220:   "anon",
	221:   "ltz",
	222:   "bitg",
	223:   "ask",
	224:   "smart",
	225:   "xuez",
	226:   "hlm",
	227:   "web",
	228:   "acm",
	229:   "nos",
	230:   "bitc",
	231:   "hth",
	232:   "tzc",
	233:   "var",
	234:   "iov",
	235:   "fio",
	236:   "bsv",
	237:   "dxn",
	238:   "qrl",
	239:   "pcx",
	240:   "loki",
	241:   "imagewallet",
	242:   "nim",
	243:   "sov",
	244:   "jct",
	245:   "slp",
	246:   "ewt",
	247:   "uc",
	248:   "exos",
	249:   "eca",
	250:   "soom",
	251:   "xrd",
	252:   "free",
	253:   "npw",
	254:   "bst",
	255:   "smartholdem",
	256:   "nano",
	257:   "btcc",
	258:   "zenprotocol",
	259:   "zest",
	260:   "abt",
	261:   "pion",
	262:   "dt3",
	263:   "zbux",
	264:   "kpl",
	265:   "tpay",
	266:   "zilla",
	267:   "ank",
	268:   "bcc",
	269:   "hpb",
	270:   "one",
	271:   "sbc",
	272:   "ipc",
	273:   "dmtc",
	274:   "ogc",
	275:   "shit",
	276:   "andes",
	277:   "arepa",
	278:   "boli",
	279:   "ril",
	280:   "htr",
	281:   "fctid",
	282:   "bravo",
	283:   "algo",
	284:   "bzx",
	285:   "gxx",
	286:   "heat",
	287:   "xdn",
	288:   "fsn",
	289:   "cpc[capricoin]",
	290:   "bold",
	291:   "iost",
	292:   "tkey",
	293:   "use",
	294:   "bcz",
	295:   "ioc",
	296:   "asf",
	297:   "mass",
	298:   "fair",
	299:   "nuko",
	300:   "gnx",
	301:   "divi",
	302:   "cmt[community]",
	303:   "euno",
	304:   "iotx",
	305:   "onion",
	306:   "8bit",
	307:   "atc",
	308:   "bts",
	309:   "ckb",
	310:   "ugas",
	311:   "ads",
	312:   "ara",
	313:   "zil",
	314:   "moac",
	315:   "swtc",
	316:   "vnsc",
	317:   "plug",
	318:   "man",
	319:   "ecc",
	320:   "rpd",
	321:   "rap",
	322:   "gard",
	323:   "zer",
	324:   "ebst",
	325:   "shard",
	326:   "linda",
	327:   "cmm",
	328:   "block",
	329:   "audax",
	330:   "luna",
	331:   "zpm",
	332:   "kuva",
	333:   "mem",
	334:   "cs",
	335:   "swift",
	336:   "fix",
	337:   "cpc[cpcchain]",
	338:   "vgo",
	339:   "dvt",
	340:   "n8v",
	341:   "mtns",
	342:   "blast",
	343:   "dct",
	344:   "aux",
	345:   "usdp",
	346:   "htdf",
	347:   "yec",
	348:   "qlc",
	349:   "tea",
	350:   "arw",
	351:   "mdm",
	352:   "cyb",
	353:   "lto",
	354:   "dot",
	355:   "aeon",
	356:   "res",
	357:   "aya",
	358:   "daps",
	359:   "csc",
	360:   "vsys",
	361:   "nollar",
	362:   "xnos",
	363:   "cpu",
	364:   "lamb",
	365:   "vct",
	366:   "czr",
	367:   "abbc",
	368:   "het",
	369:   "xas",
	370:   "vdl",
	371:   "med",
	372:   "zvc",
	373:   "vestx",
	374:   "dbt",
	375:   "seos",
	376:   "mxw",
	377:   "znz",
	378:   "xcx",
	379:   "sox",
	380:   "nyzo",
	381:   "ulc",
	382:   "ryo",
	383:   "kal",
	384:   "xsn",
	385:   "dogec",
	386:   "bmv",
	387:   "qbc",
	388:   "img",
	389:   "qos",
	390:   "pkt",
	391:   "lhd",
	392:   "cennz",
	393:   "hsn",
	394:   "cro",
	395:   "umbru",
	396:   "ton",
	397:   "near",
	398:   "xpc",
	399:   "zoc",
	400:   "nix",
	404:   "xbi",
	412:   "ain",
	416:   "slx",
	420:   "node",
	425:   "aion",
	426:   "bc",
	444:   "phr",
	447:   "din",
	457:   "ae",
	464:   "eti",
	488:   "veo",
	500:   "theta",
	501:   "sol",
	510:   "koto",
	512:   "xrd[radiant]",
	516:   "vee",
	518:   "let",
	520:   "btcv",
	526:   "bu",
	528:   "yap",
	533:   "prj",
	555:   "bcs",
	557:   "lkr",
	561:   "nty",
	600:   "ute",
	618:   "ssp",
	625:   "east",
	663:   "sfrx",
	666:   "act",
	667:   "prkl",
	668:   "ssc",
	698:   "veil",
	700:   "xdai",
	713:   "xtl",
	714:   "bnb",
	715:   "sin",
	768:   "ballz",
	777:   "btw",
	800:   "beet",
	801:   "dst[dstra]",
	808:   "qvt",
	818:   "vet",
	820:   "clo",
	831:   "cruz",
	852:   "desm",
	886:   "adf",
	888:   "neo",
	889:   "tomo",
	890:   "xsel",
	900:   "lmo",
	916:   "meta",
	966:   "matic",
	970:   "twins",
	996:   "okp",
	997:   "sum",
	998:   "lbtc",
	999:   "bcd",
	1000:  "btn",
	1001:  "tt",
	1002:  "bkt",
	1023:  "one[harmony]",
	1024:  "ont",
	1026:  "kex",
	1027:  "mcm",
	1111:  "bbc",
	1120:  "rise",
	1122:  "cmt[cybermiles]",
	1128:  "etsc",
	1145:  "cdy",
	1337:  "dfc",
	1397:  "hyc",
	1524:  "taler",
	1533:  "beam",
	1616:  "elf",
	1620:  "ath",
	1688:  "bcx",
	1729:  "xtz",
	1776:  "l-btc",
	1815:  "ada",
	1856:  "tes",
	1901:  "clc",
	1919:  "vips",
	1926:  "city",
	1977:  "xmx",
	1984:  "trtl",
	1987:  "egem",
	1989:  "hodl",
	1990:  "phl",
	1997:  "polis",
	1998:  "xmcc",
	1999:  "colx",
	2000:  "gin",
	2001:  "mnp",
	2017:  "kin",
	2018:  "eosc",
	2019:  "gbt",
	2020:  "pkc",
	2048:  "mcash",
	2049:  "true",
	2112:  "iote",
	2221:  "ask[permission]",
	2301:  "qtum",
	2302:  "etp",
	2303:  "gxc",
	2304:  "crp",
	2305:  "ela",
	2338:  "snow",
	2570:  "aoa",
	2718:  "nas",
	2894:  "reosc",
	2941:  "bnd",
	3003:  "lux",
	3030:  "xhb",
	3077:  "cos",
	3276:  "ccc",
	3377:  "roi",
	3381:  "dyn",
	3383:  "seq",
	3552:  "deo",
	3564:  "dst[destream]",
	4218:  "iota",
	4242:  "axe",
	5248:  "fic",
	5353:  "hns",
	5757:  "stacks",
	5920:  "slu",
	6060:  "go",
	6666:  "bpa",
	6688:  "safe",
	6969:  "roger",
	7777:  "btv",
	8339:  "btq",
	8888:  "sbtc",
	8964:  "nuls",
	8999:  "btp",
	9797:  "nrg",
	9888:  "btf",
	9999:  "god",
	10000: "fo",
	10291: "btr",
	11111: "ess",
	12345: "ipos",
	13107: "bty",
	13108: "ycc",
	15845: "sdgo",
	16754: "ardr",
	19165: "safe[safecoin]",
	19167: "zel",
	19169: "rito",
	20036: "xnd",
	22504: "pwr",
	25252: "bell",
	25718: "chx",
	31102: "esn",
	31337: "thepowerio",
	33416: "teo",
	33878: "btcs",
	34952: "btt",
	37992: "fxtc",
	39321: "ama",
	49344: "stash",
	// ETH reserved token range 60000-60999
	60000: "dextt.eth", // DEX test token
	// END ETH reserved token range
	65536:    "keth",
	88888:    "ryo[c0ban]",
	99999:    "wicc",
	200625:   "aka",
	200665:   "genom",
	246529:   "ats",
	424242:   "x42",
	666666:   "vite",
	1171337:  "ilt",
	1313114:  "etho",
	1313500:  "xero",
	1712144:  "lax",
	5249353:  "bco[ore]",
	5249354:  "bhd",
	5264462:  "ptn",
	5718350:  "wan",
	5741564:  "waves",
	7562605:  "sem",
	7567736:  "ion",
	7825266:  "wgr",
	7825267:  "obsr",
	61717561: "aqua",
	91927009: "kusd",
	99999998: "fluid",
	99999999: "qkc",
}
