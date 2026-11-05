package com.ravi.taga

import android.content.ContentValues
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream

class MainActivity : FlutterActivity() {
    private val CHANNEL = "com.ravi.taga/downloads"

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL).setMethodCallHandler { call, result ->
            if (call.method == "saveToPublicDownloads") {
                val sourcePath = call.argument<String>("sourcePath")
                val fileName = call.argument<String>("fileName") ?: "document.pdf"
                val mimeType = call.argument<String>("mimeType") ?: "application/pdf"

                if (sourcePath == null) {
                    result.error("INVALID_ARGUMENTS", "sourcePath cannot be null", null)
                    return@setMethodCallHandler
                }

                try {
                    val savedPath = saveFileToPublicDownloads(sourcePath, fileName, mimeType)
                    if (savedPath != null) {
                        result.success(savedPath)
                    } else {
                        result.error("SAVE_FAILED", "Failed to save file to Downloads", null)
                    }
                } catch (e: Exception) {
                    result.error("EXCEPTION", e.message, null)
                }
            } else {
                result.notImplemented()
            }
        }
    }

    private fun saveFileToPublicDownloads(sourcePath: String, fileName: String, mimeType: String): String? {
        val sourceFile = File(sourcePath)
        if (!sourceFile.exists()) return null

        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            val contentValues = ContentValues().apply {
                put(MediaStore.MediaColumns.DISPLAY_NAME, fileName)
                put(MediaStore.MediaColumns.MIME_TYPE, mimeType)
                put(MediaStore.MediaColumns.RELATIVE_PATH, Environment.DIRECTORY_DOWNLOADS)
                put(MediaStore.Downloads.IS_PENDING, 1)
            }

            val resolver = context.contentResolver
            val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, contentValues) ?: return null

            resolver.openOutputStream(uri)?.use { outputStream ->
                FileInputStream(sourceFile).use { inputStream ->
                    inputStream.copyTo(outputStream)
                }
            }

            contentValues.clear()
            contentValues.put(MediaStore.Downloads.IS_PENDING, 0)
            resolver.update(uri, contentValues, null, null)

            "${Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS).absolutePath}/$fileName"
        } else {
            val downloadsDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
            if (!downloadsDir.exists()) {
                downloadsDir.mkdirs()
            }
            val destFile = File(downloadsDir, fileName)
            FileInputStream(sourceFile).use { inputStream ->
                FileOutputStream(destFile).use { outputStream ->
                    inputStream.copyTo(outputStream)
                }
            }
            destFile.absolutePath
        }
    }
}
