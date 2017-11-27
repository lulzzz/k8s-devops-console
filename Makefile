build-react:
	sh -c "cd react; npm run build"
	cp react/build/index.html app/views/react.html
	rm -rf public/js
	mkdir -p public/js
	cp react/build/static/js/* public/js
