#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1-}" ]]; then
    echo "Usage: $0 TAG"
    echo "TAG: the tag to build or release, e.g. cli-0.0.1"
    exit 1
fi

git_tag=$1
echo "release tag: $git_tag"

# Build the release binaries for every OS/arch combination.
# It builds compressed artifacts on $release_dir.
function build_binary {
    binary_name="okctl"
    echo "build $binary_name binaries"
    version=$1

    release_dir=$2
    echo "build release artifacts to $release_dir"

    # Create tmp dir for make binaries
    mkdir -p "output"

    # Note: windows not supported yet
    platforms=("linux" "darwin")
    arch_list=("amd64" "arm64")
    for os in "${platforms[@]}"; do
        for arch in "${arch_list[@]}"; do
            echo "Building $os-$arch"
            make okctl GOOS=$os GOARCH=$arch BUILD_DIR=output/
            if [ $? -ne 0 ]; then
                echo "Build failed for $os-$arch"
                exit 1
            fi
            # Compress as tar.gz format
            tar cvfz "${release_dir}/${binary_name}_${version}_${os}_${arch}.tar.gz" -C output $binary_name
            rm output/$binary_name
        done
    done

    # Create checksum.txt
    pushd "${release_dir}"
    for release in *; do
        echo "generate checksum: $release"
        sha256sum "$release" >>checksums.txt
    done
    popd

    rmdir output
}

function create_release {
    git_tag=$1

    # This is expected to match $module.
    module=${git_tag%-*}

    # This is the version of cli
    version=${git_tag##*-}

    additional_release_artifacts_arg=""

    # Build cli binaries for all supported platforms
    if [[ "$module" == "cli" ]]; then
        release_artifact_dir=$(mktemp -d)
        build_binary "$version" "$release_artifact_dir"

        additional_release_artifacts_arg=("$release_artifact_dir"/*)

        # Create github releases
        gh release create "$git_tag" \
            --title "$git_tag" \
            --notes "$git_tag" \
            --draft "${additional_release_artifacts_arg[@]}"

        return
    fi

    # Create github releases
    gh release create "$git_tag" \
        --title "$git_tag" \
        --draft
}

create_release "$git_tag"
