#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <FTR_CLIENT_PATH>"
    exit 1
fi

FTR_CLIENT_PATH=$1

echo "Seeding categories..."

cat ./scripts/categories.txt | ./scripts/create_categories.py http://localhost:8000 > /dev/null

echo "Seeding assets..."

echo -e "1\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/0_Hair" > /dev/null
echo -e "2\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/0_Eye" > /dev/null
echo -e "3\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/1_FaceHair" > /dev/null
echo -e "4\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/2_Cloth" > /dev/null
echo -e "5\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/3_Pant" > /dev/null
echo -e "6\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/5_Armor" > /dev/null
echo -e "7\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/4_Helmet" > /dev/null
echo -e "8\npng\ny\n" | ./scripts/upload_assets.py http://localhost:8000 "$FTR_CLIENT_PATH/Assets/SPUM/Resources/Addons/Legacy/0_Unit/0_Sprite/7_Back" > /dev/null

echo "Done."
