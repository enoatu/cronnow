.PHONY: tag delete-tag retag

tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make tag TAG=vX.Y.Z"; \
		exit 1; \
	fi
	@echo "Creating tag: $(TAG)"
	git tag $(TAG)
	git push origin $(TAG)

delete-tag:
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make delete-tag TAG=vX.Y.Z"; \
		exit 1; \
	fi
	@echo "Deleting tag: $(TAG)"
	git tag -d $(TAG)
	git push --delete origin $(TAG)

retag: delete-tag tag
