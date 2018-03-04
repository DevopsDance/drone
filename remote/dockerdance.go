package remote

import (
	"github.com/drone/drone/model"
)

func DDFileBackoff(remote Remote, u *model.User, r *model.Repo, b *model.Build, f string) (out []byte, err error) {

	// let's just dump whatever we want for now before we decide where to keep it and how to make it easy to interact with
	out = []byte(`
---
pipeline:
  docker:
    image: plugins/docker
    repo: docker-registry.otwarte.xyz/${DRONE_REPO_NAME}
    tags:
      - ${DRONE_COMMIT:0:7}

  deploy:
    image: python:2
    commands:
      - pip install PyYaml jinja2 requests raven
      - curl -SsL https://raw.githubusercontent.com/bjarocki/ok-tools/master/nomad-job-builder.py -o nomad-job-builder.py
      - python nomad-job-builder.py
  `)

	return
}
