#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#main_menu
clear
echo "MAIN MENU"
echo ""
echo "Backup Database"
echo "1. Backup to Secondary"
echo "2. Backup to Development"
echo "3. Backup to Secondary and Development"
echo "4. Backup to Custom"
echo "5. Check and Copy Backup"
echo ""
echo "Backup Table"
echo "6. Backup to Secondary"
echo "7. Backup to Development"
echo "8. Backup to Secondary and Development"
echo "9. Backup to Custom"
echo ""
echo "10. Check Account Code (Other than NBC SaaS MariaDB)"
echo ""
echo "0. Exit"
read -p "Select an option: " option_0

case "$option_0" in
    0)
        exit;
        ;;
    1)
        sh backup_database_to_secondary.sh
        ;;
    2)
        sh backup_database_to_development.sh
        ;;
    3)
        sh backup_database_to_secondary_and_development.sh
        ;;
    4)
        sh backup_database_to_custom.sh
        ;;
    5)
        sh check_and_copy_backup.sh
        ;;
    6)
        sh backup_table_to_secondary.sh
        ;;
    7)
        sh backup_table_to_development.sh
        ;;
    8)
        sh backup_table_to_secondary_and_development.sh
        ;;
    9)
        sh backup_table_to_custom.sh
        ;;
    10)
        sh check_account_code.sh
        ;;
    *)
        echo "Please select a valid option";
        exit;
        ;;
esac