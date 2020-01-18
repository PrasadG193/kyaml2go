#!/bin/bash

set -o errexit
set -e

version=$(cut -d'=' -f2- .release)
if [[ -z ${version} ]]; then
    echo "Invalid version set in .release"
    exit 1
fi


if [[ -z ${GITHUB_TOKEN} ]]; then
    echo "GITHUB_TOKEN not set. Usage: GITHUB_TOKEN=<TOKEN> ./hack/release.sh"
    exit 1
fi

echo "Publishing release ${version}"

generate_changelog() {
    local version=$1

    # generate changelog from github
    github_changelog_generator infracloudio/msbotbuilder-go -t ${GITHUB_TOKEN} --future-release ${version} -o CHANGELOG.md
    sed -i '$d' CHANGELOG.md
}

publish_release() {
    local version=$1

    # create gh release
    gothub release \
	   --user infracloudio \
	   --repo msbotbuilder-go \
	   --tag $version \
	   --name "$version" \
	   --description "$version"
}

make_release() {
    local version=$1

    # tag release
    git add .release CHANGELOG.md
    git commit -m "Release $version" ;
    git tag $version ;
    git push --tags origin develop;
    echo 'Git tag pushed successfully' ;
}

generate_changelog $version
make_release $version
publish_release $version

echo "=========================== Done ============================="
echo "Congratulations!! Release ${version} published."
echo "Don't forget to add changelog in the release description."
echo "=============================================================="
