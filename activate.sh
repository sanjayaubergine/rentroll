#!/bin/bash
# chkconfig: 345 99 01
# description: activation script to start/stop Accord RentRoll
#
# processname: rentroll


HOST=localhost
PROGNAME="rentroll"
PORT=8270
WATCHDOGOPTS=""
GETFILE="/usr/local/accord/bin/getfile.sh"
RENTROLLHOME="/home/ec2-user/apps/${PROGNAME}"
DATABASENAME="${PROGNAME}"
DBUSER="ec2-user"
IAM=$(whoami)
OS=$(uname)


usage() {
    cat <<ZZEOF
RentRoll activation script.
Usage:   activate.sh [OPTIONS] CMD

This is the Accord RentRoll activation script. It is designed to work in two environments.
First, it works with Plum - Accord's test environment automation infrastructure
Second, it can work as a service script in /etc/init.d

OPTIONS:
-p port      (default is 8270)
-h hostname  (default is localhost)
-N dbname    (default is ${PROGNAME})
-T           (use this option to indicate testing rather than production)

CMD is one of: start | stop | status | restart | ready | reload | condrestart | makeprod



Examples:
Command to start ${PROGNAME}:
	bash$  activate.sh start 

Command to stop ${PROGNAME}:
	bash$  activate.sh Stop

Command to see if ${PROGNAME} is ready for commands... the response
will be "OK" if it is ready, or something else if there are problems:

    bash$  activate.sh ready
    OK
ZZEOF
}

stopwatchdog() {
	# make sure we can find it
    pidline=$(ps -ef | grep pbwatchdog | grep -v grep)
    if [ "${pidline}" != "" ]; then
        lines=$(echo "${pidline}" | wc -l)
        if [ $lines -gt "0" ]; then
            pid=$(echo "${pidline}" | awk '{print $2}')
            $(kill $pid)
        fi          
    fi      
}

#--------------------------------------------------------------
#  If we need to make this installation use the production
#  database, just invoke:
#  $ ./activate.sh makeprod
#--------------------------------------------------------------
makeProdNode() {
	${GETFILE} accord/db/confprod.json ; mv confprod.json config.json  >log.out 2>&1 
}

#--------------------------------------------------------------
#  For QA, Sandbox, and Production nodes, go through the
#  laundry list of details...
#  1. Set up permissions for the database on QA and Sandbox nodes
#  2. Install a database with some data for testing
#  3. For PDF printing, install wkhtmltopdf
#--------------------------------------------------------------
setupAppNode() {
	#---------------------
	# database
	#---------------------
	RRDB=$(echo "show databases;" | mysql | grep rentroll | wc -l)
	if [ ${RRDB} -gt "0" ]; then
	    rm -rf ${DATABASENAME}db*  >log.out 2>&1 
	    ${GETFILE} accord/db/${DATABASENAME}db.sql.gz  >log.out 2>&1 
	    gunzip ${DATABASENAME}db.sql  >log.out 2>&1 
	    echo "DROP DATABASE IF EXISTS ${DATABASENAME}; CREATE DATABASE ${DATABASENAME}; USE ${DATABASENAME};" > restore.sql
	    echo "source ${DATABASENAME}db.sql" >> restore.sql
	    echo "GRANT ALL PRIVILEGES ON ${DATABASENAME} TO 'ec2-user'@'localhost' WITH GRANT OPTION;" >> restore.sql
	    mysql ${MYSQLOPTS} < restore.sql  >log.out 2>&1 
	fi

	#---------------------
	# wkhtmltopdf
	#---------------------
	./pdfinstall.sh  >log.out 2>&1 
	
	#-----------------------------------------------------------------
	#  If no config.json exists, pull the development environment
	#  version and use it.  The Env values mean the following:
	#    0 = development environment
	#    1 = production environment
	#    2 = QA environment
	#-----------------------------------------------------------------
	if [ ! -f ./config.json ]; then
		${GETFILE} accord/db/confdev.json  >log.out 2>&1 
		mv confdev.json config.json
	fi
}

#--------------------------------------------------------------
#  For QA, Sandbox, and Production nodes, go through the
#  laundry list of details...
#  1. Set up permissions for the database on QA and Sandbox nodes
#  2. Install a database with some data for testing
#  3. For PDF printing, install wkhtmltopdf
#--------------------------------------------------------------
setupAppNode() {
	#---------------------
	# database
	#---------------------
	RRDB=$(echo "show databases;" | mysql | grep rentroll | wc -l)
	if [ ${RRDB} -lt "1" ]; then
	    rm -rf ${DATABASENAME}db*  >log.out 2>&1 
	    ${GETFILE} accord/db/${DATABASENAME}db.sql.gz  >log.out 2>&1 
	    gunzip ${DATABASENAME}db.sql  >log.out 2>&1 
	    echo "DROP DATABASE IF EXISTS ${DATABASENAME}; CREATE DATABASE ${DATABASENAME}; USE ${DATABASENAME};" > restore.sql
	    echo "source ${DATABASENAME}db.sql" >> restore.sql
	    echo "GRANT ALL PRIVILEGES ON ${DATABASENAME} TO 'ec2-user'@'localhost' WITH GRANT OPTION;" >> restore.sql
	    mysql ${MYSQLOPTS} < restore.sql  >log.out 2>&1 
	fi

	#---------------------
	# wkhtmltopdf
	#---------------------
	./pdfinstall.sh  >log.out 2>&1 
	
	#-----------------------------------------------------------------
	#  If no config.json exists, pull the development environment
	#  version and use it.  The Env values mean the following:
	#    0 = development environment
	#    1 = production environment
	#    2 = QA environment
	#-----------------------------------------------------------------
	if [ ! -f ./config.json ]; then
		${GETFILE} accord/db/confdev.json  >log.out 2>&1 
		mv confdev.json config.json
	fi
}

start() {
	# Create a database if this is a localhost instance  
	if [ ${IAM} == "root" ]; then
		x=$(grep RRDbhost config.json | grep localhost | wc -l)
	 	if (( x == 1 )); then
	 		setupAppNode
	 	fi
	fi

	if [ ${IAM} == "root" ]; then
		chown -R ec2-user *
		# chmod u+s ${PROGNAME} pbwatchdog
		# if [ $(uname) == "Linux" -a ! -f "/etc/init.d/${PROGNAME}" ]; then
		if [ $(uname) == "Linux" ]; then
			cp ./activate.sh /etc/init.d/${PROGNAME}
			chkconfig --add ${PROGNAME}
		fi
	fi

	if [ ! -f "/usr/local/share/man/man1/rentroll.1" ]; then
		./installman.sh >installman.log 2>&1

		# These are now in the source tree
		# ${GETFILE} jenkins-snapshot/rentroll/latest/rrimages.tar.gz  >log.out 2>&1 
		# tar xzvf rrimages.tar.gz  >log.out 2>&1 
		# ${GETFILE} jenkins-snapshot/rentroll/latest/rrjs.tar.gz  >log.out 2>&1 
		# tar xzvf rrjs.tar.gz  >log.out 2>&1 
		# ${GETFILE} jenkins-snapshot/rentroll/latest/fa.tar.gz  >log.out 2>&1 
		# tar xzvf fa.tar.gz  >log.out 2>&1 
	fi

	#---------------------------------------------------
	# Make sure MySQL is running, if not retry 3 times...
	#---------------------------------------------------
	i="0"
	while [ $i -lt 3 ]
	do
		i=$[$i+1]
		MSUP=$(ps -e | grep "mysqld" | wc -l)
		if [ "${MSUP}" -lt 2 ]; then
			echo "MySQL is not running. Waiting 10 sec before retry ${i}"
			sleep 10
		else
			break
		fi
	done

	if [ $i -gt 3 ]; then
		echo "[ERROR] MySQL not available after 3 retries. Aborting..."
		exit 1
	fi

	./${PROGNAME} >log.out 2>&1 &
	if [ ${IAM} == "root" ]; then
		if [ ! -d /var/run/${PROGNAME} ]; then
			mkdir /var/run/${PROGNAME}  >log.out 2>&1 
		fi
		echo $! >/var/run/${PROGNAME}/${PROGNAME}.pid
		touch /var/lock/${PROGNAME}
	fi

	#---------------------------------------------------
	# If the watchdog is NOT running, then start it...
	#---------------------------------------------------
	W=$(ps -ef | grep "rrwatchdog" | grep "bash" | wc -l)
	if [ ${W} == 0 ]; then
		./rrwatchdog &
	fi
}

stop() {
	#---------------------------------------------------
	# stop watchdog first
	#---------------------------------------------------
	W=$(ps -ef | grep "rrwatchdog" | grep "bash" | wc -l)
	if [ ${W} == 1 ]; then
		case "${OS}" in
		"Darwin")
			pid=$(ps -ef | grep rrwatchdog | grep "bash" | sed -e 's/[ \t]*[0-9][0-9]*[ \t][ \t]*\([0-9][0-9]*\)[ \t].*/\1/')
			;;
		"Linux")
			pid=$(ps -ef | grep rrwatchdog | grep "bash" | sed -e 's/[^ \t]*[ \t][ \t]*\([0-9][0-9]*\)[ \t].*/\1/')
			;;
		"*")
			echo "Unsupported Operating System"
			exit 1
			;;
		esac
		kill ${pid}
	fi

	#---------------------------------------------------
	# now stop the server
	#---------------------------------------------------
	killall -9 rentroll
	if [ ${IAM} == "root" ]; then
		sleep 6
		rm -f /var/run/${PROGNAME}/${PROGNAME}.pid /var/lock/${PROGNAME}
	fi
}

status() {
	ST=$(curl -s http://${HOST}:${PORT}/status/ | wc -c)
	case "${ST}" in
	"2")
		exit 0
		;;
	"0")
		# ${PROGNAME} is not responsive or not running.  Exit status as described in 
		# http://refspecs.linuxbase.org/LSB_3.1.0/LSB-Core-generic/LSB-Core-generic/iniscrptact.html
		if [ ${IAM} == "root" -a -f /var/run/${PROGNAME}/${PROGNAME}.pid ]; then
			exit 1
		fi
		if [ ${IAM} == "root" -a -f /var/lock/${PROGNAME} ]; then
			exit 2
		fi
		exit 3
		;;
	esac
}

reload() {
	ST=$(curl -s http://${HOST}:${PORT}/status/)
}

restart() {
	stop
	sleep 10
	start
}

while getopts ":p:qih:N:Tb" o; do
    case "${o}" in
       b)
            WATCHDOGOPTS="-b"
	    	# echo "WATCHDOGOPTS set to: ${WATCHDOGOPTS}"
            ;;
       h)
            HOST=${OPTARG}
            echo "HOST set to: ${HOST}"
            ;;
        N)
            DATABASENAME=${OPTARG}
            # echo "DATABASENAME set to: ${DATABASENAME}"
            ;;
        p)
            PORT=${OPTARG}
	    	# echo "PORT set to: ${PORT}"
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ ${OS} == "Linux" ]; then
	cd "${RENTROLLHOME}"
fi
# PBPATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd)
# cd ${PBPATH}

for arg do
	# echo '--> '"\`$arg'"
	cmd=$(echo ${arg}|tr "[:upper:]" "[:lower:]")
    case "$cmd" in
  #   "images")
		# updateImages
		# echo "Images updated"
		# ;;
	"start")
		start
		echo "OK"
		exit 0
		;;
	"stop")
		stop
		echo "OK"
		exit 0
		;;
	"ready")
		x=$(curl -s http://localhost:8270/v1/ping | grep "Accord" | wc -l)
		if (( x == 1 )); then
	        echo "OK"
			exit 0
		fi  
		echo "UNEXPECTED RESPONSE"
		exit 1
		;;  
	# "status")
	# 	status
	# 	;;
	"restart")
		restart
		echo "OK"
		exit 0
		;;
	"reload")
		reload
		exit 0
		;;
	"condrestart")
		if [ -f /var/lock/phonebook ] ; then
			restart
		fi
		;;
	"makeprod")
		makeProdNode
		;;
	"updatedb")
		setupAppNode
		;;
	*)
		echo "Unrecognized command: $arg"
		usage
		exit 1
		;;
    esac
done
