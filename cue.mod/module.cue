module: "github.com/innoai-tech/media-toolkit"

require: {
	"dagger.io":           "v0.3.0"
	"k8s.io/api":          "v0.24.1"
	"k8s.io/apimachinery": "v0.24.1"
}

require: {
	"github.com/innoai-tech/runtime": "v0.0.0-20220622091528-901d5447fbca" @indirect()
	"universe.dagger.io":             "v0.3.0"                             @indirect()
}

replace: {
	"k8s.io/api":          "" @import("go")
	"k8s.io/apimachinery": "" @import("go")
}
