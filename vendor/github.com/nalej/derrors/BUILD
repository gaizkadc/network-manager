load("@io_bazel_rules_go//go:def.bzl", "gazelle", "go_library", "go_prefix", "go_test")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")

gazelle(
    name = "gazelle",
    build_tags = [
        "darwin",
    ],
    command = "fix",
    external = "vendored",
    prefix = "github.com/nalej/derror",
)

go_prefix("github.com/nalej/derror")

go_library(
    name = "go_default_library",
    srcs = [
        "enum.go",
        "error.go",
        "interface.go",
        "test_utils.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "error_test.go",
        "json_test.go",
    ],
    library = ":go_default_library",
)
