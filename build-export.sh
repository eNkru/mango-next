#!/bin/bash
set -e

IMAGE_NAME="mango"
OUTPUT_FILE="mango-docker-image.tar.gz"

echo "Building Docker image: ${IMAGE_NAME}..."
docker build -t "${IMAGE_NAME}" -f Dockerfile .

echo "Exporting image to ${OUTPUT_FILE}..."
docker save "${IMAGE_NAME}" | gzip > "${OUTPUT_FILE}"

FILE_SIZE=$(du -h "${OUTPUT_FILE}" | cut -f1)
echo "Done! Image exported to ${OUTPUT_FILE} (${FILE_SIZE})"
