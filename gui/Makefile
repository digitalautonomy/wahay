generate-ui:
	go generate -x ui_reader.go

generate-locale:
	# This beautiful code is necessary because gotext
	# will barf if the golang.org/x/text/message package
	# exists in the vendor directory when running...
	# The world of programming is beautiful
	rm -rf ../vendor/golang.org/x/text
	go generate -x i18n.go
	git checkout ../vendor/golang.org/x/text

touch:
	ls definitions/*.xml | xargs -n1 touch

doctor: touch generate
	git diff --exit-code .
