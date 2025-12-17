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
echo "3. Restore from Secondary"
echo "4. Restore from Custom"
echo "5. Check and Restore from Daily Backup"
echo "6. Check and Restore from Last Backup"
echo "7. Check and Restore from Archive"
echo ""
echo "Backup Database"
echo "8. Backup to Custom"
echo "9. Archive (Backup and Drop)"
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
        sh /home/bin/main_menu/restore_from_secondary.sh
        ;;
    4)
        sh /home/bin/main_menu/restore_from_custom.sh
        ;;
    5)
        sh /home/bin/main_menu/check_and_restore_from_daily_backup.sh
        ;;
    6)
        sh /home/bin/main_menu/check_and_restore_from_last_backup.sh
        ;;
    7)
        sh /home/bin/main_menu/check_and_restore_from_archive.sh
        ;;
    8)
        sh /home/bin/main_menu/backup_to_custom.sh
        ;;
    9)
        sh /home/bin/main_menu/archive.sh
        ;;
    *)
        echo "Please select a valid option";
        exit;
        ;;
esac
