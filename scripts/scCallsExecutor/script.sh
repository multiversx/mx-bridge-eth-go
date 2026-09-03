#!/bin/bash
set -e

# Make script aware of its location
SCRIPTPATH="$( cd -- "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 ; pwd -P )"

source $SCRIPTPATH/config/variables.cfg
source $SCRIPTPATH/config/functions.cfg
source $SCRIPTPATH/config/menu_functions.cfg

# See if the user has passed any command line arguments and if not show the menu
if [ $# -eq 0 ]
  then

  show_menu # Show all the menu options

  COLUMNS=8
  PS3="Please select an action:"
  options=("init" "install" "upgrade" "start" "stop" "cleanup" "get_logs" "quit")

  select opt in "${options[@]}"
  do

  case $opt in

  'init')
    init
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'install')
    install
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'upgrade')
    upgrade
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'start')
    start
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'stop')
    stop
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'cleanup')
    cleanup
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'get_logs')
    get_logs
    echo -e
    read -n 1 -s -r -p "  Process finished. Press any key to continue..."
    clear
    show_menu
    ;;

  'quit')
    echo -e
    echo -e "${GREEN}---> Exiting scripts menu...${NC}"
    echo -e
    break
    ;;

  esac

  done

else

case "$1" in
'init')
  init
  ;;

'install')
  install
  ;;

'upgrade')
  upgrade
  ;;

'start')
  start
  ;;

'stop')
  stop
  ;;

'cleanup')
  cleanup
  ;;

'get_logs')
  get_logs
  ;;

esac

fi
