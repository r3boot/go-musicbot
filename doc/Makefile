GRAPHS = \
	add-new-track \
	process-using-postgres \
	search-using-bleve \
	skip-track \
	positive-rating \
	negative-rating \
	now-playing \
	synchronize \
	high-level-architecture

all: $(GRAPHS)

$(GRAPHS):
	dot -Tpng -o $@.png $@.dot