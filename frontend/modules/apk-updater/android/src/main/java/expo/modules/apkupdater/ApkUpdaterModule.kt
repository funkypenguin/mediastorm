package expo.modules.apkupdater

import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.util.Log
import androidx.core.content.FileProvider
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import java.io.File
import java.io.FileOutputStream
import java.net.HttpURLConnection
import java.net.URL

class ApkUpdaterModule : Module() {
  companion object {
    private const val TAG = "ApkUpdaterModule"
    private const val APK_DIR = "apk-updates"
    private const val MAX_REDIRECTS = 5
  }

  private val context: Context
    get() = requireNotNull(appContext.reactContext)

  override fun definition() = ModuleDefinition {
    Name("ApkUpdater")

    Events("onDownloadProgress")

    Function("cleanDownloads") {
      cleanApkDir()
    }

    AsyncFunction("downloadAndInstall") { url: String ->
      downloadAndInstallApk(url)
    }
  }

  private fun cleanApkDir() {
    val dir = File(context.cacheDir, APK_DIR)
    if (dir.exists()) {
      val count = dir.listFiles()?.size ?: 0
      dir.deleteRecursively()
      Log.d(TAG, "Cleaned $count old APK(s)")
    }
  }

  private fun downloadAndInstallApk(url: String) {
    // Clean old APKs first
    cleanApkDir()

    val dir = File(context.cacheDir, APK_DIR)
    dir.mkdirs()
    val apkFile = File(dir, "update.apk")

    try {
      downloadFile(url, apkFile)
      launchInstaller(apkFile)
    } catch (e: Exception) {
      // Clean up partial download
      apkFile.delete()
      Log.e(TAG, "Download failed: ${e.message}")
      throw e
    }
  }

  private fun downloadFile(url: String, destination: File) {
    var currentUrl = url
    var redirectCount = 0

    while (redirectCount < MAX_REDIRECTS) {
      val connection = URL(currentUrl).openConnection() as HttpURLConnection
      connection.instanceFollowRedirects = false
      connection.connectTimeout = 15_000
      connection.readTimeout = 30_000

      try {
        val responseCode = connection.responseCode

        if (responseCode in 301..303 || responseCode == 307 || responseCode == 308) {
          val location = connection.getHeaderField("Location")
            ?: throw Exception("Redirect without Location header")
          currentUrl = location
          redirectCount++
          Log.d(TAG, "Following redirect ($redirectCount) to: $location")
          connection.disconnect()
          continue
        }

        if (responseCode != 200) {
          throw Exception("HTTP $responseCode")
        }

        val totalBytes = connection.contentLength.toLong()
        var bytesDownloaded = 0L
        var lastPercent = -1

        Log.d(TAG, "Downloading APK: $totalBytes bytes from $currentUrl")

        connection.inputStream.use { input ->
          FileOutputStream(destination).use { output ->
            val buffer = ByteArray(8192)
            var bytesRead: Int

            while (input.read(buffer).also { bytesRead = it } != -1) {
              output.write(buffer, 0, bytesRead)
              bytesDownloaded += bytesRead

              if (totalBytes > 0) {
                val percent = ((bytesDownloaded * 100) / totalBytes).toInt()
                if (percent != lastPercent) {
                  lastPercent = percent
                  sendEvent("onDownloadProgress", mapOf(
                    "percent" to percent,
                    "bytesDownloaded" to bytesDownloaded,
                    "totalBytes" to totalBytes
                  ))
                }
              }
            }
          }
        }

        Log.d(TAG, "Download complete: ${destination.length()} bytes")
        return
      } finally {
        connection.disconnect()
      }
    }

    throw Exception("Too many redirects")
  }

  private fun launchInstaller(apkFile: File) {
    val authority = "${context.packageName}.apkupdater.fileprovider"
    val uri = FileProvider.getUriForFile(context, authority, apkFile)

    val intent = Intent(Intent.ACTION_VIEW).apply {
      setDataAndType(uri, "application/vnd.android.package-archive")
      addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
      addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    }

    context.startActivity(intent)
    Log.d(TAG, "Launched package installer")
  }
}
