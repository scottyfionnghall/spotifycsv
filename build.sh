#!/usr/bin/env bash

package=$1
version=$2

if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
IFS='.' read -ra fields <<< $package
package_name="spotifycsv"

platforms=("windows/amd64" "windows/386" "linux/amd64" "linux/arm64" "linux/386")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$GOOS'-'$GOARCH'-v'$version
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi	

	env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done