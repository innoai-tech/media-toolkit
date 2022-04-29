module: "github.com/innoai-tech/media-toolkit"

require: {
	"dagger.io":                      "v0.2.17-0.20220607061721-387f0ef2334f" @vcs("release-main")
	"github.com/innoai-tech/runtime": "v0.0.0-20220606084832-dc8b2154c111"
	"k8s.io/api":                     "v0.24.1"
	"k8s.io/apimachinery":            "v0.24.1"
	"universe.dagger.io":             "v0.2.17-0.20220607061721-387f0ef2334f" @vcs("release-main")
}

replace: {
	"dagger.io":          "github.com/morlay/dagger/pkg/dagger.io@release-main"
	"universe.dagger.io": "github.com/morlay/dagger/pkg/universe.dagger.io@release-main"
}

replace: {
	"k8s.io/api":          "" @import("go")
	"k8s.io/apimachinery": "" @import("go")
}
