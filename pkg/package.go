//go:generate stringer -type=TAG_ID
package rpmdb

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/xerrors"
)

type PackageInfo struct {
	Epoch     int
	Name      string
	Version   string
	Release   string
	Arch      string
	SourceRpm string
	Size      int
	License   string
	Vendor    string

	// Summary     string
	// InstallTime uint32
}

var (
	ErrNotSupport = errors.New("Not support Now")
)

type PackageInfoEx struct {
	PackageInfo
	TagsMap map[TAG_ID]interface{}
}

type TAG_ID int32
type TAG_TYPE uint32

const (
	HEADER_SIGBASE    TAG_ID = 256
	HEADER_TAGBASE    TAG_ID = 1000
	HEADER_IMAGE      TAG_ID = 61
	HEADER_SIGNATURES TAG_ID = 62
	HEADER_IMMUTABLE  TAG_ID = 63
	HEADER_REGIONS    TAG_ID = 64
	HEADER_I18NTABLE  TAG_ID = 100

	RPMTAG_HEADERIMAGE      = HEADER_IMAGE      /*!< Current image. */
	RPMTAG_HEADERSIGNATURES = HEADER_SIGNATURES /*!< Signatures. */
	RPMTAG_HEADERIMMUTABLE  = HEADER_IMMUTABLE  /*!< Original image. */
	RPMTAG_HEADERREGIONS    = HEADER_REGIONS    /*!< Regions. */

	/* Retrofit (and uniqify) signature tags for use by rpmTagGetName() and rpmQuery. */
	/* the md5 sum was broken *twice* on big endian machines */
	/* XXX 2nd underscore prevents tagTable generation */
	RPMTAG_SIG_BASE   = HEADER_SIGBASE
	RPMTAG_SIGSIZE    = RPMTAG_SIG_BASE + 1 /* i */
	RPMTAG_SIGLEMD5_1 = RPMTAG_SIG_BASE + 2 /* internal - obsolete */
	RPMTAG_SIGPGP     = RPMTAG_SIG_BASE + 3 /* x */
	RPMTAG_SIGLEMD5_2 = RPMTAG_SIG_BASE + 4 /* x internal - obsolete */
	RPMTAG_SIGMD5     = RPMTAG_SIG_BASE + 5 /* x */
	RPMTAG_PKGID      = RPMTAG_SIGMD5       /* x */
	RPMTAG_SIGGPG     = RPMTAG_SIG_BASE + 6 /* x */
	RPMTAG_SIGPGP5    = RPMTAG_SIG_BASE + 7 /* internal - obsolete */

	RPMTAG_BADSHA1_1       = RPMTAG_SIG_BASE + 8  /* internal - obsolete */
	RPMTAG_BADSHA1_2       = RPMTAG_SIG_BASE + 9  /* internal - obsolete */
	RPMTAG_PUBKEYS         = RPMTAG_SIG_BASE + 10 /* s[] */
	RPMTAG_DSAHEADER       = RPMTAG_SIG_BASE + 11 /* x */
	RPMTAG_RSAHEADER       = RPMTAG_SIG_BASE + 12 /* x */
	RPMTAG_SHA1HEADER      = RPMTAG_SIG_BASE + 13 /* s */
	RPMTAG_HDRID           = RPMTAG_SHA1HEADER    /* s */
	RPMTAG_LONGSIGSIZE     = RPMTAG_SIG_BASE + 14 /* l */
	RPMTAG_LONGARCHIVESIZE = RPMTAG_SIG_BASE + 15 /* l */
	/* RPMTAG_SIG_BASE+16 reserved */
	RPMTAG_SHA256HEADER = RPMTAG_SIG_BASE + 17 /* s */
	/* RPMTAG_SIG_BASE+18 reserved for RPMSIGTAG_FILESIGNATURES */
	/* RPMTAG_SIG_BASE+19 reserved for RPMSIGTAG_FILESIGNATURELENGTH */
	RPMTAG_VERITYSIGNATURES    = RPMTAG_SIG_BASE + 20 /* s[] */
	RPMTAG_VERITYSIGNATUREALGO = RPMTAG_SIG_BASE + 21 /* i */

	// rpmTag_e
	// ref. https://github.com/rpm-software-management/rpm/blob/rpm-4.11.3-release/lib/rpmtag.h#L28
	RPMTAG_NAME               TAG_ID = 1000                   /* s */
	RPMTAG_N                         = RPMTAG_NAME            /* s */
	RPMTAG_VERSION            TAG_ID = 1001                   /* s */
	RPMTAG_V                         = RPMTAG_VERSION         /* s */
	RPMTAG_RELEASE            TAG_ID = 1002                   /* s */
	RPMTAG_R                         = RPMTAG_RELEASE         /* s */
	RPMTAG_EPOCH              TAG_ID = 1003                   /* i */
	RPMTAG_E                         = RPMTAG_EPOCH           /* i */
	RPMTAG_SUMMARY            TAG_ID = 1004                   /* s{} */
	RPMTAG_DESCRIPTION        TAG_ID = 1005                   /* s{} */
	RPMTAG_BUILDTIME          TAG_ID = 1006                   /* i */
	RPMTAG_BUILDHOST          TAG_ID = 1007                   /* s */
	RPMTAG_INSTALLTIME        TAG_ID = 1008                   /* i */
	RPMTAG_SIZE               TAG_ID = 1009                   /* i */
	RPMTAG_DISTRIBUTION       TAG_ID = 1010                   /* s */
	RPMTAG_VENDOR             TAG_ID = 1011                   /* s */
	RPMTAG_GIF                TAG_ID = 1012                   /* x */
	RPMTAG_XPM                TAG_ID = 1013                   /* x */
	RPMTAG_LICENSE            TAG_ID = 1014                   /* s */
	RPMTAG_PACKAGER           TAG_ID = 1015                   /* s */
	RPMTAG_GROUP              TAG_ID = 1016                   /* s{} */
	RPMTAG_CHANGELOG          TAG_ID = 1017                   /* s[] internal */
	RPMTAG_SOURCE             TAG_ID = 1018                   /* s[] */
	RPMTAG_PATCH              TAG_ID = 1019                   /* s[] */
	RPMTAG_URL                TAG_ID = 1020                   /* s */
	RPMTAG_OS                 TAG_ID = 1021                   /* s legacy used int */
	RPMTAG_ARCH               TAG_ID = 1022                   /* s legacy used int */
	RPMTAG_PREIN              TAG_ID = 1023                   /* s */
	RPMTAG_POSTIN             TAG_ID = 1024                   /* s */
	RPMTAG_PREUN              TAG_ID = 1025                   /* s */
	RPMTAG_POSTUN             TAG_ID = 1026                   /* s */
	RPMTAG_OLDFILENAMES       TAG_ID = 1027                   /* s[] obsolete */
	RPMTAG_FILESIZES          TAG_ID = 1028                   /* i[] */
	RPMTAG_FILESTATES         TAG_ID = 1029                   /* c[] */
	RPMTAG_FILEMODES          TAG_ID = 1030                   /* h[] */
	RPMTAG_FILEUIDS           TAG_ID = 1031                   /* i[] internal - obsolete */
	RPMTAG_FILEGIDS           TAG_ID = 1032                   /* i[] internal - obsolete */
	RPMTAG_FILERDEVS          TAG_ID = 1033                   /* h[] */
	RPMTAG_FILEMTIMES         TAG_ID = 1034                   /* i[] */
	RPMTAG_FILEDIGESTS        TAG_ID = 1035                   /* s[] */
	RPMTAG_FILEMD5S                  = RPMTAG_FILEDIGESTS     /* s[] */
	RPMTAG_FILELINKTOS        TAG_ID = 1036                   /* s[] */
	RPMTAG_FILEFLAGS          TAG_ID = 1037                   /* i[] */
	RPMTAG_ROOT               TAG_ID = 1038                   /* internal - obsolete */
	RPMTAG_FILEUSERNAME       TAG_ID = 1039                   /* s[] */
	RPMTAG_FILEGROUPNAME      TAG_ID = 1040                   /* s[] */
	RPMTAG_EXCLUDE            TAG_ID = 1041                   /* internal - obsolete */
	RPMTAG_EXCLUSIVE          TAG_ID = 1042                   /* internal - obsolete */
	RPMTAG_ICON               TAG_ID = 1043                   /* x */
	RPMTAG_SOURCERPM          TAG_ID = 1044                   /* s */
	RPMTAG_FILEVERIFYFLAGS    TAG_ID = 1045                   /* i[] */
	RPMTAG_ARCHIVESIZE        TAG_ID = 1046                   /* i */
	RPMTAG_PROVIDENAME        TAG_ID = 1047                   /* s[] */
	RPMTAG_PROVIDES                  = RPMTAG_PROVIDENAME     /* s[] */
	RPMTAG_P                         = RPMTAG_PROVIDENAME     /* s[] */
	RPMTAG_REQUIREFLAGS       TAG_ID = 1048                   /* i[] */
	RPMTAG_REQUIRENAME        TAG_ID = 1049                   /* s[] */
	RPMTAG_REQUIRES           TAG_ID = RPMTAG_REQUIRENAME     /* s[] */
	RPMTAG_REQUIREVERSION     TAG_ID = 1050                   /* s[] */
	RPMTAG_NOSOURCE           TAG_ID = 1051                   /* i[] */
	RPMTAG_NOPATCH            TAG_ID = 1052                   /* i[] */
	RPMTAG_CONFLICTFLAGS      TAG_ID = 1053                   /* i[] */
	RPMTAG_CONFLICTNAME       TAG_ID = 1054                   /* s[] */
	RPMTAG_CONFLICTS                 = RPMTAG_CONFLICTNAME    /* s[] */
	RPMTAG_C                         = RPMTAG_CONFLICTNAME    /* s[] */
	RPMTAG_CONFLICTVERSION    TAG_ID = 1055                   /* s[] */
	RPMTAG_DEFAULTPREFIX      TAG_ID = 1056                   /* s internal - deprecated */
	RPMTAG_BUILDROOT          TAG_ID = 1057                   /* s internal - obsolete */
	RPMTAG_INSTALLPREFIX      TAG_ID = 1058                   /* s internal - deprecated */
	RPMTAG_EXCLUDEARCH        TAG_ID = 1059                   /* s[] */
	RPMTAG_EXCLUDEOS          TAG_ID = 1060                   /* s[] */
	RPMTAG_EXCLUSIVEARCH      TAG_ID = 1061                   /* s[] */
	RPMTAG_EXCLUSIVEOS        TAG_ID = 1062                   /* s[] */
	RPMTAG_AUTOREQPROV        TAG_ID = 1063                   /* s internal */
	RPMTAG_RPMVERSION         TAG_ID = 1064                   /* s */
	RPMTAG_TRIGGERSCRIPTS     TAG_ID = 1065                   /* s[] */
	RPMTAG_TRIGGERNAME        TAG_ID = 1066                   /* s[] */
	RPMTAG_TRIGGERVERSION     TAG_ID = 1067                   /* s[] */
	RPMTAG_TRIGGERFLAGS       TAG_ID = 1068                   /* i[] */
	RPMTAG_TRIGGERINDEX       TAG_ID = 1069                   /* i[] */
	RPMTAG_VERIFYSCRIPT       TAG_ID = 1079                   /* s */
	RPMTAG_CHANGELOGTIME      TAG_ID = 1080                   /* i[] */
	RPMTAG_CHANGELOGNAME      TAG_ID = 1081                   /* s[] */
	RPMTAG_CHANGELOGTEXT      TAG_ID = 1082                   /* s[] */
	RPMTAG_BROKENMD5          TAG_ID = 1083                   /* internal - obsolete */
	RPMTAG_PREREQ             TAG_ID = 1084                   /* internal */
	RPMTAG_PREINPROG          TAG_ID = 1085                   /* s[] */
	RPMTAG_POSTINPROG         TAG_ID = 1086                   /* s[] */
	RPMTAG_PREUNPROG          TAG_ID = 1087                   /* s[] */
	RPMTAG_POSTUNPROG         TAG_ID = 1088                   /* s[] */
	RPMTAG_BUILDARCHS         TAG_ID = 1089                   /* s[] */
	RPMTAG_OBSOLETENAME       TAG_ID = 1090                   /* s[] */
	RPMTAG_OBSOLETES                 = RPMTAG_OBSOLETENAME    /* s[] */
	RPMTAG_O                         = RPMTAG_OBSOLETENAME    /* s[] */
	RPMTAG_VERIFYSCRIPTPROG   TAG_ID = 1091                   /* s[] */
	RPMTAG_TRIGGERSCRIPTPROG  TAG_ID = 1092                   /* s[] */
	RPMTAG_DOCDIR             TAG_ID = 1093                   /* internal */
	RPMTAG_COOKIE             TAG_ID = 1094                   /* s */
	RPMTAG_FILEDEVICES        TAG_ID = 1095                   /* i[] */
	RPMTAG_FILEINODES         TAG_ID = 1096                   /* i[] */
	RPMTAG_FILELANGS          TAG_ID = 1097                   /* s[] */
	RPMTAG_PREFIXES           TAG_ID = 1098                   /* s[] */
	RPMTAG_INSTPREFIXES       TAG_ID = 1099                   /* s[] */
	RPMTAG_TRIGGERIN          TAG_ID = 1100                   /* internal */
	RPMTAG_TRIGGERUN          TAG_ID = 1101                   /* internal */
	RPMTAG_TRIGGERPOSTUN      TAG_ID = 1102                   /* internal */
	RPMTAG_AUTOREQ            TAG_ID = 1103                   /* internal */
	RPMTAG_AUTOPROV           TAG_ID = 1104                   /* internal */
	RPMTAG_CAPABILITY         TAG_ID = 1105                   /* i internal - obsolete */
	RPMTAG_SOURCEPACKAGE      TAG_ID = 1106                   /* i */
	RPMTAG_OLDORIGFILENAMES   TAG_ID = 1107                   /* internal - obsolete */
	RPMTAG_BUILDPREREQ        TAG_ID = 1108                   /* internal */
	RPMTAG_BUILDREQUIRES      TAG_ID = 1109                   /* internal */
	RPMTAG_BUILDCONFLICTS     TAG_ID = 1110                   /* internal */
	RPMTAG_BUILDMACROS        TAG_ID = 1111                   /* internal - unused */
	RPMTAG_PROVIDEFLAGS       TAG_ID = 1112                   /* i[] */
	RPMTAG_PROVIDEVERSION     TAG_ID = 1113                   /* s[] */
	RPMTAG_OBSOLETEFLAGS      TAG_ID = 1114                   /* i[] */
	RPMTAG_OBSOLETEVERSION    TAG_ID = 1115                   /* s[] */
	RPMTAG_DIRINDEXES         TAG_ID = 1116                   /* i[] */
	RPMTAG_BASENAMES          TAG_ID = 1117                   /* s[] */
	RPMTAG_DIRNAMES           TAG_ID = 1118                   /* s[] */
	RPMTAG_ORIGDIRINDEXES     TAG_ID = 1119                   /* i[] relocation */
	RPMTAG_ORIGBASENAMES      TAG_ID = 1120                   /* s[] relocation */
	RPMTAG_ORIGDIRNAMES       TAG_ID = 1121                   /* s[] relocation */
	RPMTAG_OPTFLAGS           TAG_ID = 1122                   /* s */
	RPMTAG_DISTURL            TAG_ID = 1123                   /* s */
	RPMTAG_PAYLOADFORMAT      TAG_ID = 1124                   /* s */
	RPMTAG_PAYLOADCOMPRESSOR  TAG_ID = 1125                   /* s */
	RPMTAG_PAYLOADFLAGS       TAG_ID = 1126                   /* s */
	RPMTAG_INSTALLCOLOR       TAG_ID = 1127                   /* i transaction color when installed */
	RPMTAG_INSTALLTID         TAG_ID = 1128                   /* i */
	RPMTAG_REMOVETID          TAG_ID = 1129                   /* i */
	RPMTAG_SHA1RHN            TAG_ID = 1130                   /* internal - obsolete */
	RPMTAG_RHNPLATFORM        TAG_ID = 1131                   /* s internal - obsolete */
	RPMTAG_PLATFORM           TAG_ID = 1132                   /* s */
	RPMTAG_PATCHESNAME        TAG_ID = 1133                   /* s[] deprecated placeholder (SuSE) */
	RPMTAG_PATCHESFLAGS       TAG_ID = 1134                   /* i[] deprecated placeholder (SuSE) */
	RPMTAG_PATCHESVERSION     TAG_ID = 1135                   /* s[] deprecated placeholder (SuSE) */
	RPMTAG_CACHECTIME         TAG_ID = 1136                   /* i internal - obsolete */
	RPMTAG_CACHEPKGPATH       TAG_ID = 1137                   /* s internal - obsolete */
	RPMTAG_CACHEPKGSIZE       TAG_ID = 1138                   /* i internal - obsolete */
	RPMTAG_CACHEPKGMTIME      TAG_ID = 1139                   /* i internal - obsolete */
	RPMTAG_FILECOLORS         TAG_ID = 1140                   /* i[] */
	RPMTAG_FILECLASS          TAG_ID = 1141                   /* i[] */
	RPMTAG_CLASSDICT          TAG_ID = 1142                   /* s[] */
	RPMTAG_FILEDEPENDSX       TAG_ID = 1143                   /* i[] */
	RPMTAG_FILEDEPENDSN       TAG_ID = 1144                   /* i[] */
	RPMTAG_DEPENDSDICT        TAG_ID = 1145                   /* i[] */
	RPMTAG_SOURCEPKGID        TAG_ID = 1146                   /* x */
	RPMTAG_FILECONTEXTS       TAG_ID = 1147                   /* s[] - obsolete */
	RPMTAG_FSCONTEXTS         TAG_ID = 1148                   /* s[] extension */
	RPMTAG_RECONTEXTS         TAG_ID = 1149                   /* s[] extension */
	RPMTAG_POLICIES           TAG_ID = 1150                   /* s[] selinux *.te policy file. */
	RPMTAG_PRETRANS           TAG_ID = 1151                   /* s */
	RPMTAG_POSTTRANS          TAG_ID = 1152                   /* s */
	RPMTAG_PRETRANSPROG       TAG_ID = 1153                   /* s[] */
	RPMTAG_POSTTRANSPROG      TAG_ID = 1154                   /* s[] */
	RPMTAG_DISTTAG            TAG_ID = 1155                   /* s */
	RPMTAG_OLDSUGGESTSNAME    TAG_ID = 1156                   /* s[] - obsolete */
	RPMTAG_OLDSUGGESTS               = RPMTAG_OLDSUGGESTSNAME /* s[] - obsolete */
	RPMTAG_OLDSUGGESTSVERSION TAG_ID = 1157                   /* s[] - obsolete */
	RPMTAG_OLDSUGGESTSFLAGS   TAG_ID = 1158                   /* i[] - obsolete */
	RPMTAG_OLDENHANCESNAME    TAG_ID = 1159                   /* s[] - obsolete */
	RPMTAG_OLDENHANCES               = RPMTAG_OLDENHANCESNAME /* s[] - obsolete */
	RPMTAG_OLDENHANCESVERSION TAG_ID = 1160                   /* s[] - obsolete */
	RPMTAG_OLDENHANCESFLAGS   TAG_ID = 1161                   /* i[] - obsolete */
	RPMTAG_PRIORITY           TAG_ID = 1162                   /* i[] extension placeholder (unimplemented) */
	RPMTAG_CVSID              TAG_ID = 1163                   /* s (unimplemented) */
	RPMTAG_SVNID                     = RPMTAG_CVSID           /* s (unimplemented) */
	RPMTAG_BLINKPKGID         TAG_ID = 1164                   /* s[] (unimplemented) */
	RPMTAG_BLINKHDRID         TAG_ID = 1165                   /* s[] (unimplemented) */
	RPMTAG_BLINKNEVRA         TAG_ID = 1166                   /* s[] (unimplemented) */
	RPMTAG_FLINKPKGID         TAG_ID = 1167                   /* s[] (unimplemented) */
	RPMTAG_FLINKHDRID         TAG_ID = 1168                   /* s[] (unimplemented) */
	RPMTAG_FLINKNEVRA         TAG_ID = 1169                   /* s[] (unimplemented) */
	RPMTAG_PACKAGEORIGIN      TAG_ID = 1170                   /* s (unimplemented) */
	RPMTAG_TRIGGERPREIN       TAG_ID = 1171                   /* internal */
	RPMTAG_BUILDSUGGESTS      TAG_ID = 1172                   /* internal (unimplemented) */
	RPMTAG_BUILDENHANCES      TAG_ID = 1173                   /* internal (unimplemented) */
	RPMTAG_SCRIPTSTATES       TAG_ID = 1174                   /* i[] scriptlet exit codes (unimplemented) */
	RPMTAG_SCRIPTMETRICS      TAG_ID = 1175                   /* i[] scriptlet execution times (unimplemented) */
	RPMTAG_BUILDCPUCLOCK      TAG_ID = 1176                   /* i (unimplemented) */
	RPMTAG_FILEDIGESTALGOS    TAG_ID = 1177                   /* i[] (unimplemented) */
	RPMTAG_VARIANTS           TAG_ID = 1178                   /* s[] (unimplemented) */
	RPMTAG_XMAJOR             TAG_ID = 1179                   /* i (unimplemented) */
	RPMTAG_XMINOR             TAG_ID = 1180                   /* i (unimplemented) */
	RPMTAG_REPOTAG            TAG_ID = 1181                   /* s (unimplemented) */
	RPMTAG_KEYWORDS           TAG_ID = 1182                   /* s[] (unimplemented) */
	RPMTAG_BUILDPLATFORMS     TAG_ID = 1183                   /* s[] (unimplemented) */
	RPMTAG_PACKAGECOLOR       TAG_ID = 1184                   /* i (unimplemented) */
	RPMTAG_PACKAGEPREFCOLOR   TAG_ID = 1185                   /* i (unimplemented) */
	RPMTAG_XATTRSDICT         TAG_ID = 1186                   /* s[] (unimplemented) */
	RPMTAG_FILEXATTRSX        TAG_ID = 1187                   /* i[] (unimplemented) */
	RPMTAG_DEPATTRSDICT       TAG_ID = 1188                   /* s[] (unimplemented) */
	RPMTAG_CONFLICTATTRSX     TAG_ID = 1189                   /* i[] (unimplemented) */
	RPMTAG_OBSOLETEATTRSX     TAG_ID = 1190                   /* i[] (unimplemented) */
	RPMTAG_PROVIDEATTRSX      TAG_ID = 1191                   /* i[] (unimplemented) */
	RPMTAG_REQUIREATTRSX      TAG_ID = 1192                   /* i[] (unimplemented) */
	RPMTAG_BUILDPROVIDES      TAG_ID = 1193                   /* internal (unimplemented) */
	RPMTAG_BUILDOBSOLETES     TAG_ID = 1194                   /* internal (unimplemented) */
	RPMTAG_DBINSTANCE         TAG_ID = 1195                   /* i extension */
	RPMTAG_NVRA               TAG_ID = 1196                   /* s extension */

	/* tags 1997-4999 reserved */
	RPMTAG_FILENAMES                   TAG_ID = 5000                  /* s[] extension */
	RPMTAG_FILEPROVIDE                 TAG_ID = 5001                  /* s[] extension */
	RPMTAG_FILEREQUIRE                 TAG_ID = 5002                  /* s[] extension */
	RPMTAG_FSNAMES                     TAG_ID = 5003                  /* s[] (unimplemented) */
	RPMTAG_FSSIZES                     TAG_ID = 5004                  /* l[] (unimplemented) */
	RPMTAG_TRIGGERCONDS                TAG_ID = 5005                  /* s[] extension */
	RPMTAG_TRIGGERTYPE                 TAG_ID = 5006                  /* s[] extension */
	RPMTAG_ORIGFILENAMES               TAG_ID = 5007                  /* s[] extension */
	RPMTAG_LONGFILESIZES               TAG_ID = 5008                  /* l[] */
	RPMTAG_LONGSIZE                    TAG_ID = 5009                  /* l */
	RPMTAG_FILECAPS                    TAG_ID = 5010                  /* s[] */
	RPMTAG_FILEDIGESTALGO              TAG_ID = 5011                  /* i file digest algorithm */
	RPMTAG_BUGURL                      TAG_ID = 5012                  /* s */
	RPMTAG_EVR                         TAG_ID = 5013                  /* s extension */
	RPMTAG_NVR                         TAG_ID = 5014                  /* s extension */
	RPMTAG_NEVR                        TAG_ID = 5015                  /* s extension */
	RPMTAG_NEVRA                       TAG_ID = 5016                  /* s extension */
	RPMTAG_HEADERCOLOR                 TAG_ID = 5017                  /* i extension */
	RPMTAG_VERBOSE                     TAG_ID = 5018                  /* i extension */
	RPMTAG_EPOCHNUM                    TAG_ID = 5019                  /* i extension */
	RPMTAG_PREINFLAGS                  TAG_ID = 5020                  /* i */
	RPMTAG_POSTINFLAGS                 TAG_ID = 5021                  /* i */
	RPMTAG_PREUNFLAGS                  TAG_ID = 5022                  /* i */
	RPMTAG_POSTUNFLAGS                 TAG_ID = 5023                  /* i */
	RPMTAG_PRETRANSFLAGS               TAG_ID = 5024                  /* i */
	RPMTAG_POSTTRANSFLAGS              TAG_ID = 5025                  /* i */
	RPMTAG_VERIFYSCRIPTFLAGS           TAG_ID = 5026                  /* i */
	RPMTAG_TRIGGERSCRIPTFLAGS          TAG_ID = 5027                  /* i[] */
	RPMTAG_COLLECTIONS                 TAG_ID = 5029                  /* s[] list of collections (unimplemented) */
	RPMTAG_POLICYNAMES                 TAG_ID = 5030                  /* s[] */
	RPMTAG_POLICYTYPES                 TAG_ID = 5031                  /* s[] */
	RPMTAG_POLICYTYPESINDEXES          TAG_ID = 5032                  /* i[] */
	RPMTAG_POLICYFLAGS                 TAG_ID = 5033                  /* i[] */
	RPMTAG_VCS                         TAG_ID = 5034                  /* s */
	RPMTAG_ORDERNAME                   TAG_ID = 5035                  /* s[] */
	RPMTAG_ORDERVERSION                TAG_ID = 5036                  /* s[] */
	RPMTAG_ORDERFLAGS                  TAG_ID = 5037                  /* i[] */
	RPMTAG_MSSFMANIFEST                TAG_ID = 5038                  /* s[] reservation (unimplemented) */
	RPMTAG_MSSFDOMAIN                  TAG_ID = 5039                  /* s[] reservation (unimplemented) */
	RPMTAG_INSTFILENAMES               TAG_ID = 5040                  /* s[] extension */
	RPMTAG_REQUIRENEVRS                TAG_ID = 5041                  /* s[] extension */
	RPMTAG_PROVIDENEVRS                TAG_ID = 5042                  /* s[] extension */
	RPMTAG_OBSOLETENEVRS               TAG_ID = 5043                  /* s[] extension */
	RPMTAG_CONFLICTNEVRS               TAG_ID = 5044                  /* s[] extension */
	RPMTAG_FILENLINKS                  TAG_ID = 5045                  /* i[] extension */
	RPMTAG_RECOMMENDNAME               TAG_ID = 5046                  /* s[] */
	RPMTAG_RECOMMENDS                         = RPMTAG_RECOMMENDNAME  /* s[] */
	RPMTAG_RECOMMENDVERSION            TAG_ID = 5047                  /* s[] */
	RPMTAG_RECOMMENDFLAGS              TAG_ID = 5048                  /* i[] */
	RPMTAG_SUGGESTNAME                 TAG_ID = 5049                  /* s[] */
	RPMTAG_SUGGESTS                           = RPMTAG_SUGGESTNAME    /* s[] */
	RPMTAG_SUGGESTVERSION              TAG_ID = 5050                  /* s[] extension */
	RPMTAG_SUGGESTFLAGS                TAG_ID = 5051                  /* i[] extension */
	RPMTAG_SUPPLEMENTNAME              TAG_ID = 5052                  /* s[] */
	RPMTAG_SUPPLEMENTS                        = RPMTAG_SUPPLEMENTNAME /* s[] */
	RPMTAG_SUPPLEMENTVERSION           TAG_ID = 5053                  /* s[] */
	RPMTAG_SUPPLEMENTFLAGS             TAG_ID = 5054                  /* i[] */
	RPMTAG_ENHANCENAME                 TAG_ID = 5055                  /* s[] */
	RPMTAG_ENHANCES                           = RPMTAG_ENHANCENAME    /* s[] */
	RPMTAG_ENHANCEVERSION              TAG_ID = 5056                  /* s[] */
	RPMTAG_ENHANCEFLAGS                TAG_ID = 5057                  /* i[] */
	RPMTAG_RECOMMENDNEVRS              TAG_ID = 5058                  /* s[] extension */
	RPMTAG_SUGGESTNEVRS                TAG_ID = 5059                  /* s[] extension */
	RPMTAG_SUPPLEMENTNEVRS             TAG_ID = 5060                  /* s[] extension */
	RPMTAG_ENHANCENEVRS                TAG_ID = 5061                  /* s[] extension */
	RPMTAG_ENCODING                    TAG_ID = 5062                  /* s */
	RPMTAG_FILETRIGGERIN               TAG_ID = 5063                  /* internal */
	RPMTAG_FILETRIGGERUN               TAG_ID = 5064                  /* internal */
	RPMTAG_FILETRIGGERPOSTUN           TAG_ID = 5065                  /* internal */
	RPMTAG_FILETRIGGERSCRIPTS          TAG_ID = 5066                  /* s[] */
	RPMTAG_FILETRIGGERSCRIPTPROG       TAG_ID = 5067                  /* s[] */
	RPMTAG_FILETRIGGERSCRIPTFLAGS      TAG_ID = 5068                  /* i[] */
	RPMTAG_FILETRIGGERNAME             TAG_ID = 5069                  /* s[] */
	RPMTAG_FILETRIGGERINDEX            TAG_ID = 5070                  /* i[] */
	RPMTAG_FILETRIGGERVERSION          TAG_ID = 5071                  /* s[] */
	RPMTAG_FILETRIGGERFLAGS            TAG_ID = 5072                  /* i[] */
	RPMTAG_TRANSFILETRIGGERIN          TAG_ID = 5073                  /* internal */
	RPMTAG_TRANSFILETRIGGERUN          TAG_ID = 5074                  /* internal */
	RPMTAG_TRANSFILETRIGGERPOSTUN      TAG_ID = 5075                  /* internal */
	RPMTAG_TRANSFILETRIGGERSCRIPTS     TAG_ID = 5076                  /* s[] */
	RPMTAG_TRANSFILETRIGGERSCRIPTPROG  TAG_ID = 5077                  /* s[] */
	RPMTAG_TRANSFILETRIGGERSCRIPTFLAGS TAG_ID = 5078                  /* i[] */
	RPMTAG_TRANSFILETRIGGERNAME        TAG_ID = 5079                  /* s[] */
	RPMTAG_TRANSFILETRIGGERINDEX       TAG_ID = 5080                  /* i[] */
	RPMTAG_TRANSFILETRIGGERVERSION     TAG_ID = 5081                  /* s[] */
	RPMTAG_TRANSFILETRIGGERFLAGS       TAG_ID = 5082                  /* i[] */
	RPMTAG_REMOVEPATHPOSTFIXES         TAG_ID = 5083                  /* s internal */
	RPMTAG_FILETRIGGERPRIORITIES       TAG_ID = 5084                  /* i[] */
	RPMTAG_TRANSFILETRIGGERPRIORITIES  TAG_ID = 5085                  /* i[] */
	RPMTAG_FILETRIGGERCONDS            TAG_ID = 5086                  /* s[] extension */
	RPMTAG_FILETRIGGERTYPE             TAG_ID = 5087                  /* s[] extension */
	RPMTAG_TRANSFILETRIGGERCONDS       TAG_ID = 5088                  /* s[] extension */
	RPMTAG_TRANSFILETRIGGERTYPE        TAG_ID = 5089                  /* s[] extension */
	RPMTAG_FILESIGNATURES              TAG_ID = 5090                  /* s[] */
	RPMTAG_FILESIGNATURELENGTH         TAG_ID = 5091                  /* i */
	RPMTAG_PAYLOADDIGEST               TAG_ID = 5092                  /* s[] */
	RPMTAG_PAYLOADDIGESTALGO           TAG_ID = 5093                  /* i */
	RPMTAG_AUTOINSTALLED               TAG_ID = 5094                  /* i reservation (unimplemented) */
	RPMTAG_IDENTITY                    TAG_ID = 5095                  /* s reservation (unimplemented) */
	RPMTAG_MODULARITYLABEL             TAG_ID = 5096                  /* s */
	RPMTAG_PAYLOADDIGESTALT            TAG_ID = 5097                  /* s[] */

	//rpmTagType_e
	// ref. https://github.com/rpm-software-management/rpm/blob/rpm-4.11.3-release/lib/rpmtag.h#L362
	RPM_NULL_TYPE         TAG_TYPE = 0
	RPM_CHAR_TYPE         TAG_TYPE = 1
	RPM_INT8_TYPE         TAG_TYPE = 2
	RPM_INT16_TYPE        TAG_TYPE = 3
	RPM_INT32_TYPE        TAG_TYPE = 4
	RPM_INT64_TYPE        TAG_TYPE = 5
	RPM_STRING_TYPE       TAG_TYPE = 6
	RPM_BIN_TYPE          TAG_TYPE = 7
	RPM_STRING_ARRAY_TYPE TAG_TYPE = 8
	RPM_I18NSTRING_TYPE   TAG_TYPE = 9
)

func dumpEntry(entry *indexEntry) error {
	reader := bytes.NewReader(entry.Data)
	switch entry.Info.Type {
	case RPM_NULL_TYPE:
	case RPM_CHAR_TYPE, RPM_INT8_TYPE:
		var value byte
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return xerrors.Errorf("failed to read binary byte: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return xerrors.Errorf("failed to read binary byte: %w", err)
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)

	case RPM_INT16_TYPE:
		var value uint16
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return xerrors.Errorf("failed to read binary byte: %w", err)
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)

	case RPM_INT32_TYPE:
		var value uint32
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return xerrors.Errorf("failed to read binary byte: %w", err)
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)

	case RPM_INT64_TYPE:
		var value uint64
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return xerrors.Errorf("failed to read binary byte: %w", err)
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)

	case RPM_STRING_TYPE:
		value := string(bytes.TrimRight(entry.Data, "\x00"))
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)

	case RPM_BIN_TYPE:
		if entry.Info.Tag >= RPMTAG_HEADERIMAGE && entry.Info.Tag < RPMTAG_HEADERREGIONS {

		} else {
			value := hex.EncodeToString(entry.Data[:entry.Info.Count])
			fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, value)
		}

	case RPM_STRING_ARRAY_TYPE:
		var values = make([]string, entry.Info.Count)
		subStrings := bytes.SplitN(entry.Data, []byte("\x00"), int(entry.Info.Count))
		for i := 0; i < len(values) && i < len(subStrings); i++ {
			values[i] = string(subStrings[i])
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, values)
	case RPM_I18NSTRING_TYPE:
		var values = make([]string, entry.Info.Count)
		subStrings := bytes.SplitN(entry.Data, []byte("\x00"), int(entry.Info.Count))
		for i := 0; i < len(values) && i < len(subStrings); i++ {
			values[i] = string(subStrings[i])
		}
		fmt.Printf("TAG: %v, TYPE: %v, DATA: %v\n", entry.Info.Tag, entry.Info.Type, values)
	}
	return nil
}

func entryValue(entry *indexEntry) (interface{}, error) {
	reader := bytes.NewReader(entry.Data)
	switch entry.Info.Type {
	case RPM_NULL_TYPE:
	case RPM_CHAR_TYPE, RPM_INT8_TYPE:
		var value byte
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return nil, xerrors.Errorf("failed to read binary byte: %w", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return nil, xerrors.Errorf("failed to read binary byte: %w", err)
		}
		return value, nil

	case RPM_INT16_TYPE:
		var value uint16
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return nil, xerrors.Errorf("failed to read binary byte: %w", err)
		}
		return value, nil

	case RPM_INT32_TYPE:
		var value uint32
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return nil, xerrors.Errorf("failed to read binary byte: %w", err)
		}
		return value, nil

	case RPM_INT64_TYPE:
		var value uint64
		if err := binary.Read(reader, binary.BigEndian, &value); err != nil {
			return nil, xerrors.Errorf("failed to read binary byte: %w", err)
		}
		return value, nil

	case RPM_STRING_TYPE:
		value := string(bytes.TrimRight(entry.Data, "\x00"))
		return value, nil

	case RPM_BIN_TYPE:
		if entry.Info.Tag >= RPMTAG_HEADERIMAGE && entry.Info.Tag < RPMTAG_HEADERREGIONS {
			//TODO:
		} else {
			value := hex.EncodeToString(entry.Data[:entry.Info.Count])
			return value, nil
		}

	case RPM_STRING_ARRAY_TYPE:
		var values = make([]string, entry.Info.Count)
		subStrings := bytes.SplitN(entry.Data, []byte("\x00"), int(entry.Info.Count))
		for i := 0; i < len(values) && i < len(subStrings); i++ {
			values[i] = string(subStrings[i])
		}
		return values, nil

	case RPM_I18NSTRING_TYPE:
		var values = make([]string, entry.Info.Count)
		subStrings := bytes.SplitN(entry.Data, []byte("\x00"), int(entry.Info.Count))
		for i := 0; i < len(values) && i < len(subStrings); i++ {
			values[i] = string(subStrings[i])
		}
		return values, nil
	}
	return nil, ErrNotSupport
}

// ref. https://github.com/rpm-software-management/rpm/blob/rpm-4.11.3-release/lib/tagexts.c#L649
func getNEVRA(indexEntries []indexEntry) (*PackageInfo, error) {
	pkgInfo := &PackageInfo{}

	for _, indexEntry := range indexEntries {
		// dumpEntry(&indexEntry)
		// fmt.Printf("TAG: %v, TYPE: %v, len=%v\n", indexEntry.Info.Tag, indexEntry.Info.Type, indexEntry.Info.Count)
		switch indexEntry.Info.Tag {
		case RPMTAG_NAME:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag name")
			}
			pkgInfo.Name = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_EPOCH:
			if indexEntry.Info.Type != RPM_INT32_TYPE {
				return nil, xerrors.New("invalid tag epoch")
			}

			var epoch int32
			reader := bytes.NewReader(indexEntry.Data)
			if err := binary.Read(reader, binary.BigEndian, &epoch); err != nil {
				return nil, xerrors.Errorf("failed to read binary (epoch): %w", err)
			}
			pkgInfo.Epoch = int(epoch)
		case RPMTAG_VERSION:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag version")
			}
			pkgInfo.Version = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_RELEASE:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag release")
			}
			pkgInfo.Release = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_ARCH:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag arch")
			}
			pkgInfo.Arch = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_SOURCERPM:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag sourcerpm")
			}
			pkgInfo.SourceRpm = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.SourceRpm == "(none)" {
				pkgInfo.SourceRpm = ""
			}
		case RPMTAG_LICENSE:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag license")
			}
			pkgInfo.License = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.License == "(none)" {
				pkgInfo.License = ""
			}
		case RPMTAG_VENDOR:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag vendor")
			}
			pkgInfo.Vendor = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.Vendor == "(none)" {
				pkgInfo.Vendor = ""
			}
		case RPMTAG_SIZE:
			if indexEntry.Info.Type != RPM_INT32_TYPE {
				return nil, xerrors.New("invalid tag size")
			}

			var size int32
			reader := bytes.NewReader(indexEntry.Data)
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return nil, xerrors.Errorf("failed to read binary (size): %w", err)
			}
			pkgInfo.Size = int(size)
		}
	}
	//fmt.Printf("===PKG: %v\n", pkgInfo.Name)
	return pkgInfo, nil
}

func getPackageWithTags(indexEntries []indexEntry, tagMask map[TAG_ID]bool) (*PackageInfoEx, error) {
	pkgInfo := &PackageInfoEx{}
	pkgInfo.TagsMap = make(map[TAG_ID]interface{})

	for _, indexEntry := range indexEntries {
		switch indexEntry.Info.Tag {
		case RPMTAG_NAME:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag name")
			}
			pkgInfo.Name = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_EPOCH:
			if indexEntry.Info.Type != RPM_INT32_TYPE {
				return nil, xerrors.New("invalid tag epoch")
			}

			var epoch int32
			reader := bytes.NewReader(indexEntry.Data)
			if err := binary.Read(reader, binary.BigEndian, &epoch); err != nil {
				return nil, xerrors.Errorf("failed to read binary (epoch): %w", err)
			}
			pkgInfo.Epoch = int(epoch)
		case RPMTAG_VERSION:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag version")
			}
			pkgInfo.Version = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_RELEASE:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag release")
			}
			pkgInfo.Release = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_ARCH:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag arch")
			}
			pkgInfo.Arch = string(bytes.TrimRight(indexEntry.Data, "\x00"))
		case RPMTAG_SOURCERPM:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag sourcerpm")
			}
			pkgInfo.SourceRpm = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.SourceRpm == "(none)" {
				pkgInfo.SourceRpm = ""
			}
		case RPMTAG_LICENSE:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag license")
			}
			pkgInfo.License = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.License == "(none)" {
				pkgInfo.License = ""
			}
		case RPMTAG_VENDOR:
			if indexEntry.Info.Type != RPM_STRING_TYPE {
				return nil, xerrors.New("invalid tag vendor")
			}
			pkgInfo.Vendor = string(bytes.TrimRight(indexEntry.Data, "\x00"))
			if pkgInfo.Vendor == "(none)" {
				pkgInfo.Vendor = ""
			}

		case RPMTAG_SIZE:
			if indexEntry.Info.Type != RPM_INT32_TYPE {
				return nil, xerrors.New("invalid tag size")
			}

			var size int32
			reader := bytes.NewReader(indexEntry.Data)
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return nil, xerrors.Errorf("failed to read binary (size): %w", err)
			}
			pkgInfo.Size = int(size)
		default:
			if tagMask[indexEntry.Info.Tag] {
				if v, err := entryValue(&indexEntry); err == nil {
					pkgInfo.TagsMap[indexEntry.Info.Tag] = v
				}
			}
		}
	}
	return pkgInfo, nil
}
