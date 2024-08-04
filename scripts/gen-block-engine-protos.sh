#!/usr/bin/env bash

# Define the go_package base path
GO_PACKAGE_BASE="github.com/Prophet-Solutions/block-engine-protos"

# Define the go package directory
GO_PACKAGE_DIR="block-engine-pb"

# Directory containing .proto files
PROTO_DIR="mev-protos"

which protoc || { echo "protoc not found"; exit 1; }

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

export PATH=$PATH:$(go env GOPATH)/bin

chmod +r $PROTO_DIR/*.proto

# Function to add go_package to a .proto file
add_go_package() {
    local proto_file=$1
    local package_name=$(basename "$proto_file" .proto)
    local go_package_line="option go_package = \"$GO_PACKAGE_BASE/$package_name\";"

    # Check if go_package line already exists
    if grep -q "^option go_package" "$proto_file"; then
        if grep -q "^option go_package = \"$GO_PACKAGE_BASE/$package_name\";" "$proto_file"; then
            echo "go_package option is already correct in $proto_file. Skipping."
            return
        else
            echo "Updating go_package option in $proto_file."
            # Update the existing go_package line
            sed -i.bak "s|^option go_package = .*|$go_package_line|" "$proto_file"
            return
        fi
    fi

    # Read the file and insert go_package line after package declaration
    awk -v pkg_line="$go_package_line" '
    {
        print $0
        if ($1 == "package") {
            print pkg_line
        }
    }' "$proto_file" > "$proto_file.tmp" && mv "$proto_file.tmp" "$proto_file"
}

# Iterate over all .proto files in the directory
for proto_file in "$PROTO_DIR"/*.proto; 
do
    add_go_package "$proto_file"
done

protoc \
    --go_out=. \
    --go-grpc_out=. \
    --proto_path=$PROTO_DIR \
    $PROTO_DIR/*.proto

# Create the go package directory if it does not exist
mkdir -p $GO_PACKAGE_DIR

# Find and copy directories containing .pb.go files
find "$GO_PACKAGE_BASE" -name "*.pb.go" -exec sh -c 'cp -r "$(dirname "{}")" "$0"' "$GO_PACKAGE_DIR" \;

# Extract the first part before the first slash
BASE_FOLDER=$(echo "$GO_PACKAGE_BASE" | cut -d'/' -f1)

# Remove the folder
rm -rf "$BASE_FOLDER"

cd $GO_PACKAGE_DIR

# Initialize the go module if it does not exist
if [ ! -f go.mod ]; then
    go mod init $GO_PACKAGE_BASE
fi

go mod tidy
