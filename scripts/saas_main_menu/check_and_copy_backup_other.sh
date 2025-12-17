#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"
zones=("aws1_saas_mariadb" "aws3_saas_mariadb" "huawei_saas_mariadb" "nbc_saas_sqlserver")

check() {
    cluster="${account_detail[0]}"
    main_database="${account_detail[1]}"
    zone="${account_detail[2]}"
    dmart_database="${main_database}_dmart"
    databases=("$main_database" "$dmart_database")
    # echo ""
    # echo -e "Zone\t\t: $zone"
    # echo -e "Cluster\t\t: $cluster"
    # echo -e "Database\t: $main_database"
    case "$zone" in
    "aws1_saas_mariadb")
        source_path="/media/ArchiveDB_SFAsia"
        ;;
    "aws3_saas_mariadb")
        source_path="/media/ArchiveDB_SF7/aws_3/$cluster"
        ;;
    "huawei_saas_mariadb")
        source_path="/media/ArchiveDB_SF7/humatrix/$cluster"
        ;;
    "nbc_saas_sqlserver")
        source_path="/media/ArchiveDB_SFID/$cluster"
        ;;
    esac
    if [[ -z "$source_path" ]]
    then
        echo "Please provide a valid source path";
        exit;
    fi
    echo ""
    echo "Please wait while searching for available backup on $cluster"
    echo ""
    full_file_find=$(find ${source_path}/*/*/*/${main_database}/ -type f -name "\\[FULL\\]_${main_database}_[0-9]*.sql.enc" -o -name "${main_database}_[0-9]*.bak" 2>/dev/null | sort -n);
    if [[ $(echo "$full_file_find" | wc -l)>0 ]] && [[ -n "$full_file_find" ]]
    then
        echo "Available backup files of $account_code:"
        readarray full_file_arrays <<< "$full_file_find"
        full_file_number="1"
        for full_file_array in ${full_file_arrays[@]}
        do
            echo "$full_file_number. $(date '+%d %B %Y' -r $full_file_array)"
            full_file_number="$(($full_file_number+1))"
        done
        echo ""
        echo "0. Exit"
        sleep 1
        read -p "Select the date of backup files to restore: " date_chosen
        echo ""
        if [[ -z "$date_chosen" ]]
        then
            echo "Please select a valid date";
            exit;
        elif [[ "$date_chosen" == "0" ]]
        then
            exit
        elif [[ "$date_chosen" -gt 0 && "$date_chosen" -lt "$full_file_number" ]]
        then
            full_file_index="$(($date_chosen-1))"
            full_file_chosen="$(echo -n "${full_file_arrays[full_file_index]}")"
            if [[ -f "$full_file_chosen" ]] && [[ -n "$full_file_chosen" ]]
            then
                clear
                echo "CHECK AND COPY BACKUP"
                echo ""
                echo "Availabe backup files of $account_code on $(date '+%d %B %Y' -r $full_file_chosen):"
                backup_file_find=$(ls -rt $(dirname "$full_file_chosen") | grep -v "\[LOG\]_NEW")
                readarray backup_file_arrays <<< "$backup_file_find"
                column_output=""
                backup_file_number="1"
                formatted_time_index="0"
                formatted_time_arrays=()
                for backup_file_array in ${backup_file_arrays[@]}
                do
                    timestamp=$(echo $backup_file_array | sed 's/\.sql\.enc$//' | rev | cut -d'_' -f1-2 | rev)
                    date_part=$(echo $timestamp | cut -d'_' -f1)
                    time_part=$(echo $timestamp | cut -d'_' -f2)
                    date_string="${date_part}${time_part}"
                    formatted_time=$(date -d "${date_part:0:4}-${date_part:4:2}-${date_part:6:2} ${time_part:0:2}:${time_part:2:2}:${time_part:4:2}" "+%d %B %Y %H:%M:%S")
                    formatted_time_arrays+=("$formatted_time")
                    if [[ "$backup_file_array" == *\[FULL\]* ]]
                    then
                        column_output="$column_output$backup_file_number. $formatted_time [FULL]\n"
                    elif [[ "$backup_file_array" == *.bak ]]
                    then
                        column_output="$column_output$backup_file_number. $formatted_time [FULL]\n"
                    else
                        column_output="$column_output$backup_file_number. $formatted_time\n"
                    fi
                    backup_file_number="$(($backup_file_number+1))"
                done
                echo -e "$column_output" | column
                echo ""
                echo "0. Exit"
                sleep 1
                read -p "Select backup file to be copied: " option_backup
                echo ""
                if [[ $option_backup -gt 0 ]]
                then
                    if [[ "${backup_file_arrays[$(($option_backup-1))]}" == *\[FULL\]* ]]
                    then
                        backup_file_type="1"
                        # copy
                        echo "Sorry, not yet available, please wait for the next update ;)";
                        exit;
                    else
                        echo "Sorry, not yet available, please wait for the next update ;)";
                        exit;
                    fi
                elif [[ $option_backup == "0" ]]
                then
                    exit
                else
                    echo "Please select a valid backup file";
                    exit;
                fi
            else
            echo "No backup available. Please check manually"
            fi
        else
            echo "Please select a valid date";
            exit;
        fi
    else
        echo "No backup available"
    fi
}

clear
echo "CHECK ACCOUNT CODE (Other than in NBC SaaS MariaDB)"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
if [[ -z "$account_code" ]]
then
    echo "Please provide a valid account_code";
    exit;
fi
for zone in "${zones[@]}"; do
    eval "${zone}_account_detail=($(mysql -u"$user" -p"$password" -h172.17.70.170 -P3306 -D dbsf_tools -sNe "SELECT vip, dbname FROM dba_all_backup_db_${zone} WHERE dbname LIKE CONCAT('%', '$account_code') AND DATE(tanggal) = CURDATE() - INTERVAL 1 DAY ORDER BY tanggal DESC LIMIT 1;") '$zone')"
done
account_detail=("${aws1_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${aws3_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${huawei_saas_mariadb_account_detail[@]}")
[ ${#account_detail[@]} -eq 1 ] && account_detail=("${nbc_saas_sqlserver_account_detail[@]}")
if [ "${#account_detail[@]}" -gt 1 ]; then
    check
else
    echo "Account code not found"
fi
