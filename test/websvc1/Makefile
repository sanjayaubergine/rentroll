all:
	@echo "*** Completed in websvc1 ***"

clean:
	rm -f rentroll.log log llog *.g ./gold/*.g err.txt [a-z] [a-z][0-9] [a-z0-9][0-9][0-9] qq? fail request w.[a-z]
	@echo "*** CLEAN completed in svc ***"

test:
	@touch fail
	./functest.sh
	@echo "*** TEST completed in websvc1 ***"
	@rm -f fail

package:
	@echo "*** PACKAGE completed in websvc1 ***"

get:
	@touch fail
	../../tools/bashtools/getdb.sh
	@rm -f fail
