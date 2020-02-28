package internal

const (
	FluxBanner  = "Flux-go // fast gateway for microservice: dubbo, grpc, http"
	FluxVersion = "Version // git.commit=%s, build.version=%s, build.date=%s"
)

type BuildVersion struct {
	CommitId string
	Version  string
	Date     string
}
