#!/bin/sh
snap_install() {
	name=$1
	if [ -n "$SNAP_CHANNEL" ] ; then
		# Don't reinstall if we have it installed already
		if ! snap list | grep $name ; then
			snap install --$SNAP_CHANNEL $name
		fi
	else
		snap install --dangerous $PROJECT_PATH/$name*_amd64.snap
	fi
}

# waits for a service to be active. Besides that, waits enough
# time after detecting it to be active to prevent restarting
# same service too quickly several times.
# $1 service name
# $2 start limit interval in seconds (default set to 10. Exec `systemctl show THE_SERVICE -p StartLimitInterval` to verify)
# $3 start limit burst. Times allowed to start the service in start_limit_interval time (default set to 5. Exec `systemctl show THE_SERVICE -p StartLimitBurst` to verify)
wait_for_systemd_service() {
	while ! systemctl status $1 ; do
		sleep 1
	done
  # As debian services default limit is to allow 5 restarts in a 10sec period
  # (StartLimitInterval=10000000 and StartLimitBurst=5), having enough wait time we
  # prevent "service: Start request repeated too quickly" error.
  #
  # You can check those values for certain service by typing:
  #   $systemctl show THE_SERVICE -p StartLimitInterval,StartLimitBurst
  #
  if [ $# -ge 2 ]; then
    start_limit_interval = $2
  else
    start_limit_interval=$(systemctl show $1 -p StartLimitInterval | sed 's/StartLimitInterval=//')
    # original limit interval is provided in microseconds.
    start_limit_interval=$((start_limit_interval / 1000000))
  fi

  if [ $# -eq 3 ]; then
    start_limit_burst = $3
  else
    start_limit_burst=$(systemctl show $1 -p StartLimitBurst | sed 's/StartLimitBurst=//')
  fi

  # adding 1 to be sure we exceed the limit
  sleep_time=$((1 + $start_limit_interval / $start_limit_burst))  
	sleep $sleep_time
}

wait_for_serial_vault() {
	wait_for_systemd_service snap.serial-vault-server.serial-vault.service
}

stop_after_first_reboot() {
	if [ $SPREAD_REBOOT -eq 1 ] ; then
		exit 0
	fi
}

# $1 instruction to execute repeatedly until complete or max times
# $2 sleep time between retries. Default 1sec
# $3 max_iterations. Default 20
repeat_until_done() {
  timeout=1
  if [ $# -ge 2 ]; then
    timeout=$2
  fi

  max_iterations=20
  if [ $# -ge 3 ]; then
    max_iterations=$3
  fi

  i=0
  while [ $i -lt $max_iterations ] ; do
      if $(eval $1) ; then
          break
      fi
      sleep $timeout
      let i=i+1
  done
  test $i -lt $max_iterations
}
