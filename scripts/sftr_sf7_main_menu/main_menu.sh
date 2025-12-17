#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#main_menu
clear
echo "MAIN MENU"
echo ""
echo "Restore Database"
echo "1. Restore from Production (Live)"
echo "2. Restore from Production (Custom Date and Time)"
echo "3. Check and Restore from Daily Backup"
echo "4. Check and Restore from Last Backup"
echo "5. Check and Restore from Archive"
echo ""
echo "Backup Database"
echo "6. Backup to Development"
echo "7. Backup to Custom"
echo "8. Archive (Backup and Drop)"
echo ""
echo "0. Exit"
read -p "Select an option: " option_0

case "$option_0" in
    0)
        exit;
        ;;
    1)
        sh /home/bin/main_menu/restore_from_production_live.sh
        ;;
    2)
        sh /home/bin/main_menu/restore_from_production_custom_date_and_time.sh
        ;;
    3)
        sh /home/bin/main_menu/check_and_restore_from_daily_backup.sh
        ;;
    4)
        sh /home/bin/main_menu/check_and_restore_from_last_backup.sh
        ;;
    5)
        sh /home/bin/main_menu/check_and_restore_from_archive.sh
        ;;
    6)
        sh /home/bin/main_menu/backup_to_development.sh
        ;;
    7)
        sh /home/bin/main_menu/backup_to_custom.sh
        ;;
    8)
        sh /home/bin/main_menu/archive.sh
        ;;
    *)
        echo "Please select a valid option";
        exit;
        ;;
esac
