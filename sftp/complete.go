package sftp

import "github.com/chzyer/readline"

//var completer = readline.NewPrefixCompleter(
//	readline.PcItem("mode",
//		readline.PcItem("vi"),
//		readline.PcItem("emacs"),
//	),
//	readline.PcItem("login"),
//	readline.PcItem("say",
//		readline.PcItemDynamic(listFiles("./"),
//			readline.PcItem("with",
//				readline.PcItem("following"),
//				readline.PcItem("items"),
//			),
//		),
//		readline.PcItem("hello"),
//		readline.PcItem("bye"),
//	),
//	readline.PcItem("setprompt"),
//	readline.PcItem("setpassword"),
//	readline.PcItem("bye"),
//	readline.PcItem("help"),
//	readline.PcItem("go",
//		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
//		readline.PcItem("install",
//			readline.PcItem("-v"),
//			readline.PcItem("-vv"),
//			readline.PcItem("-vvv"),
//		),
//		readline.PcItem("test"),
//	),
//	readline.PcItem("sleep"),
//)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("get",
		readline.PcItem("-r"),
	),
	readline.PcItem("put",
		readline.PcItem("-r"),
	),
	readline.PcItem("cd"),
	readline.PcItem("lcd"),
)
