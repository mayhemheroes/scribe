// Package drone contains types that represent the Drone configuration schema for generation.
// This is an example of a Drone pipeline:
//
//	kind: pipeline
// 	type: docker
// 	name: basic pipeline
// 	clone:
// 	  disable: true
// 	steps:
// 	- name: clone
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=0 ./demo/basic
// 	- name: install-frontend-dependencies
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=1 ./demo/basic
// 	  depends_on:
// 	  - clone
// 	- name: install-backend-dependencies
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=2 ./demo/basic
// 	  depends_on:
// 	  - clone
// 	- name: write-version-file
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=3 ./demo/basic
// 	  depends_on:
// 	  - install-frontend-dependencies
// 	  - install-backend-dependencies
// 	- name: compile-backend
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=4 ./demo/basic
// 	  depends_on:
// 	  - write-version-file
// 	- name: compile-frontend
// 	  image: grafana/shipwright:latest
// 	  commands:
// 	  - shipwright -step=5 ./demo/basic
// 	  depends_on:
// 	  - compile-backend
package drone
