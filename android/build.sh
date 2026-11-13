set -e
cd "$(dirname "$0")"
JAR=${ANDROID_JAR:-/usr/lib/android-sdk/platforms/android-23/android.jar}
rm -rf build/classes build/*.apk build/classes.dex; mkdir -p build/classes
javac -nowarn --release 8 -cp $JAR -d build/classes src/br/observatorio/limeira/MainActivity.java 2>&1 | grep -v "warning\|JAVA_TOOL" || true
dalvik-exchange --dex --min-sdk-version=23 --output=build/classes.dex build/classes
aapt package -f -M AndroidManifest.xml -S res -A assets -I $JAR -F build/u.apk --min-sdk-version 23 --target-sdk-version 34 --version-code 1 --version-name 1.0
(cd build && aapt add u.apk classes.dex >/dev/null)
zipalign -f 4 build/u.apk build/a.apk
[ -f build/release.jks ] || keytool -genkeypair -keystore build/release.jks -storepass obslimeira -keypass obslimeira -alias obs -keyalg RSA -keysize 2048 -validity 10000 -dname "CN=Observatorio Limeira, L=Limeira, ST=SP, C=BR" 2>/dev/null
apksigner sign --ks build/release.jks --ks-pass pass:obslimeira --key-pass pass:obslimeira --v1-signing-enabled true --v2-signing-enabled true --out build/ObservatorioLimeira.apk build/a.apk 2>&1 | grep -v JAVA_TOOL || true
apksigner verify --verbose --min-sdk-version 23 build/ObservatorioLimeira.apk 2>&1 | grep -v JAVA_TOOL | head -6
aapt dump badging build/ObservatorioLimeira.apk | head -8
unzip -l build/ObservatorioLimeira.apk; ls -la build/*.apk
