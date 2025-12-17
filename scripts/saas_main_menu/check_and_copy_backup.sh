#!/bin/bash
#Created by Muhammad Naufal Saniar
#set -x

#global_variables
year=`date +%Y`; month=`date +%m`; day=`date +%d`; hour=`date +%H`; minute=`date +%M`; second=`date +%S`;
user="sst_user"
password="demo"

copy() {
    echo "1. Copy to Secondary"
    echo "2. Copy to Development"
    echo "3. Copy to Secondary and Development"
    echo ""
    echo "0. Exit"
    read -p "Select the destination: " destination_option
    echo ""
    case "$app" in
    "GD"|"GD-PRO"|"SF-6")
        if [[ "$vip" == "v168" ]]
        then
            app_check=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -D "$main_database" -sNe "SELECT pversion FROM tsfclicense;")
            if [[ "$app_check" == "6.0" ]]
            then
                path_secondary="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
                path_development="/media/dbtolocal-nbc/sf6_gdpro/production/${year}${month}${day}"
            elif [[ "$app_check" == "7.0" || "$app_check" == "8.0" ]]
            then
                path_secondary="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
                path_development="/media/dbtolocal-nbc/sf7/production/${year}${month}${day}"
            fi
        else
            path_secondary="/media/dbtosecondary-nbc/sf6_gdpro/production/${year}${month}${day}"
            path_development="/media/dbtolocal-nbc/sf6_gdpro/production/${year}${month}${day}"
        fi
        ;;
    "SF-7"|"SF7-NBC2")
        path_secondary="/media/dbtosecondary-nbc/sf7/production/${year}${month}${day}"
        path_development="/media/dbtolocal-nbc/sf7/production/${year}${month}${day}"
        ;;
    esac
    if [[ -z "$path_secondary" ]] || [[ -z "$path_development" ]]
    then
        echo "Please provide a valid destination path";
        exit;
    fi
    destinations=()
    case "$destination_option" in
        0)
            exit;
            ;;
        1)
            destinations=("$path_secondary")
            ;;
        2)
            destinations=("$path_development")
            ;;
        3)
            destinations=("$path_secondary" "$path_development")
            ;;
        *)
            echo "Please select a valid destination";
            exit;
            ;;
    esac
    for destination in "${destinations[@]}"
    do
        mkdir -p "$destination"
        if [[ "$destination" == "$path_secondary" ]]
        then
            grant="${destination}/grant_${account_code}.sql"
            > $grant
            grant_users=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SELECT user FROM mysql.user WHERE user LIKE CONCAT('%$account_code','_admin') OR user LIKE CONCAT('%$account_code','_user') OR user LIKE CONCAT('%$account_code','_fin')")
            for grant_user in ${grant_users[@]}
            do
                grant_password=$(mysql -u"$user" -p"$password" -h"$host" -P"$port" -sNe "SELECT password FROM mysql.user WHERE user='$grant_user';")
                echo "GRANT ALL PRIVILEGES ON database.* TO '$grant_user'@'%' IDENTIFIED BY PASSWORD '$grant_password';" >> "$grant"
            done
        fi
        if [[ "$backup_file_type" == "full" ]]
        then
            # cp --preserve=timestamps "$full_file_chosen" "$destination"
            # cp --preserve=timestamps "${full_file_chosen//${account_code}/${account_code}_dmart}" "$destination"
            cp --preserve=timestamps "$backup_file_chosen" "$destination"
            cp --preserve=timestamps "$dmart_backup_file_chosen" "$destination"
            echo "Backup of $account_code on ${formatted_time_arrays[0]} copied to $destination"
        fi
    done
}

check() {
    main_database=$(echo "${account_detail[0]}" | tr '[:upper:]' '[:lower:]')
    vip="${account_detail[1]}"
    host="${account_detail[2]}"
    port="${account_detail[3]}"
    app="${account_detail[4]}"
    dmart_database="${main_database}_dmart"
    databases=("$main_database" "$dmart_database")
    case "$app" in
    "GD"|"GD-PRO")
        source_path="/media/ArchiveDB_GDPRO"
        ;;
    "SF-6")
        source_path="/media/ArchiveDB_SF6"
        ;;
    "SF-7"|"SF7-NBC2")
        source_path="/media/ArchiveDB_SF7"
        ;;
    esac
    if [[ -z "$source_path" ]]
    then
        echo "Please provide a valid source path";
        exit;
    fi
    echo ""
    echo "Please wait while searching for available backup on $vip"
    echo ""
    full_file_find=$(find ${source_path}/${vip}/*/*/*/${main_database}/ -type f -name "\\[FULL\\]_${main_database}_[0-9]*.sql.enc" 2>/dev/null | sort -n);
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
                        backup_file_type="full"
                        backup_file_chosen="$(dirname "$full_file_chosen")/$(echo -n "${backup_file_arrays[$((option_backup-1))]}" | tr -d '\n')"
                        dmart_backup_file_chosen="${backup_file_chosen//${account_code}/${account_code}_dmart}"
                        copy
                    # elif [[ "${backup_file_arrays[$(($option_backup-1))]}" == *\[LOG\]* ]]
                    # then
                        # backup_file_type="incremental"
                        # backup_file_chosen_arrays=($(printf "$(dirname "$full_file_chosen")/%s" "${backup_file_arrays[@]:0:$option_backup}"))
                        # echo "${backup_file_chosen_arrays[@]}"
                        # copy
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
echo "CHECK AND COPY BACKUP"
echo ""
read -p "Enter account code: " account_code
account_code=$(echo "$account_code" | tr '[:upper:]' '[:lower:]')
account_detail=($(mysql -u"$user" -p"$password" -h172.17.70.170 -P3306 -D dbsf_tools -sNe "SELECT A.dbname, B.vip, B.ip, B.port_1, B.sf FROM tmshdata A INNER JOIN v_dba_master_cluster B ON (A.host = B.virtual_ip AND A.port = B.port_2 AND B.is_master = 'Master') WHERE A.codename = '$account_code';"))
if [[ -z "${account_detail[*]}" ]]
then
    echo "Account detail not found";
    exit;
fi
check
