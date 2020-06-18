EXE=ts3

run:
	go build -o $(EXE)
	./$(EXE)

install:
	go build -o $(EXE)
	mv $(EXE) /bin/$(EXE)

