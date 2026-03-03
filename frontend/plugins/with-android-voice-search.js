const { withAndroidManifest, withDangerousMod } = require('@expo/config-plugins');
const fs = require('fs');
const path = require('path');

/**
 * Add voice search support for Android TV.
 *
 * When the user presses the microphone button on an Android TV remote (e.g.
 * NVIDIA Shield), the system performs speech recognition and sends the result
 * as an ACTION_SEARCH intent. This plugin:
 *
 * 1. Adds an ACTION_SEARCH intent filter + searchable.xml metadata to the manifest
 * 2. Creates the searchable.xml resource
 * 3. Modifies MainActivity to convert voice search intents into deep link URLs
 *    so React Native / expo-router can route them to the search screen
 */

const withVoiceSearchManifest = (config) => {
  return withAndroidManifest(config, (config) => {
    const manifest = config.modResults.manifest;
    const application = manifest.application?.[0];

    if (!application?.activity) {
      console.warn('\u26a0\ufe0f [VoiceSearch] No activities found in manifest');
      return config;
    }

    const mainActivity = application.activity.find(
      (activity) => activity.$?.['android:name'] === '.MainActivity'
    );

    if (!mainActivity) {
      console.warn('\u26a0\ufe0f [VoiceSearch] MainActivity not found');
      return config;
    }

    // Add SEARCH intent filter
    if (!mainActivity['intent-filter']) {
      mainActivity['intent-filter'] = [];
    }

    const hasSearchFilter = mainActivity['intent-filter'].some((filter) =>
      filter.action?.some(
        (action) => action.$?.['android:name'] === 'android.intent.action.SEARCH'
      )
    );

    if (!hasSearchFilter) {
      mainActivity['intent-filter'].push({
        action: [{ $: { 'android:name': 'android.intent.action.SEARCH' } }],
      });
    }

    // Add searchable metadata to activity
    if (!mainActivity['meta-data']) {
      mainActivity['meta-data'] = [];
    }

    const hasSearchable = mainActivity['meta-data'].some(
      (meta) => meta.$?.['android:name'] === 'android.app.searchable'
    );

    if (!hasSearchable) {
      mainActivity['meta-data'].push({
        $: {
          'android:name': 'android.app.searchable',
          'android:resource': '@xml/searchable',
        },
      });
    }

    console.log('\u2705 [VoiceSearch] Added search intent filter to AndroidManifest.xml');
    return config;
  });
};

const withVoiceSearchResources = (config) => {
  return withDangerousMod(config, [
    'android',
    async (config) => {
      const xmlDir = path.join(
        config.modRequest.platformProjectRoot,
        'app/src/main/res/xml'
      );

      if (!fs.existsSync(xmlDir)) {
        fs.mkdirSync(xmlDir, { recursive: true });
      }

      const searchablePath = path.join(xmlDir, 'searchable.xml');
      const searchableContent = `<?xml version="1.0" encoding="utf-8"?>
<searchable xmlns:android="http://schemas.android.com/apk/res/android"
    android:label="@string/app_name"
    android:hint="Search for movies or TV shows"
    android:voiceSearchMode="showVoiceSearchButton|launchRecognizer" />
`;

      fs.writeFileSync(searchablePath, searchableContent);
      console.log('\u2705 [VoiceSearch] Created searchable.xml');

      return config;
    },
  ]);
};

const withVoiceSearchMainActivity = (config) => {
  return withDangerousMod(config, [
    'android',
    async (config) => {
      const mainActivityPath = path.join(
        config.modRequest.platformProjectRoot,
        'app/src/main/java/com/strmr/app/MainActivity.kt'
      );

      if (!fs.existsSync(mainActivityPath)) {
        console.warn(
          '\u26a0\ufe0f [VoiceSearch] MainActivity.kt not found at:',
          mainActivityPath
        );
        return config;
      }

      let content = fs.readFileSync(mainActivityPath, 'utf-8');

      if (content.includes('ACTION_SEARCH')) {
        console.log('\u2139\ufe0f [VoiceSearch] MainActivity already has voice search code');
        return config;
      }

      // Add required imports
      const importsToAdd = [
        'import android.app.SearchManager',
        'import android.content.Intent',
      ];

      for (const importLine of importsToAdd) {
        if (!content.includes(importLine)) {
          const lastImportMatch = content.match(/^import .+$/gm);
          if (lastImportMatch) {
            const lastImport = lastImportMatch[lastImportMatch.length - 1];
            content = content.replace(lastImport, `${lastImport}\n${importLine}`);
          }
        }
      }

      // Add voice search intent conversion to onCreate (handles cold start)
      const onCreateMatch = content.match(
        /override fun onCreate\(savedInstanceState: Bundle\?\) \{/
      );
      if (onCreateMatch) {
        const insertPoint =
          content.indexOf(onCreateMatch[0]) + onCreateMatch[0].length;
        const voiceSearchOnCreate = `
    // Convert voice search intent to deep link before React Native processes it
    android.util.Log.d("VoiceSearch", "onCreate called, intent action: \${intent?.action}, extras: \${intent?.extras?.keySet()?.joinToString()}")
    intent?.let {
      if (it.action == Intent.ACTION_SEARCH) {
        val query = it.getStringExtra(SearchManager.QUERY)
        android.util.Log.d("VoiceSearch", "onCreate ACTION_SEARCH received, query: $query")
        if (!query.isNullOrBlank()) {
          val uri = android.net.Uri.parse("com.strmr.app:///search?q=\${android.net.Uri.encode(query)}")
          android.util.Log.d("VoiceSearch", "onCreate converting to deep link: $uri")
          setIntent(Intent(Intent.ACTION_VIEW, uri))
        } else {
          android.util.Log.w("VoiceSearch", "onCreate ACTION_SEARCH but query was null/blank")
        }
      } else {
        android.util.Log.d("VoiceSearch", "onCreate intent is not ACTION_SEARCH, action: \${it.action}")
      }
    }`;
        content =
          content.slice(0, insertPoint) +
          voiceSearchOnCreate +
          content.slice(insertPoint);
      }

      // Add onNewIntent override (handles warm start — app already running)
      const onNewIntentCode = `
  override fun onNewIntent(intent: Intent) {
    android.util.Log.d("VoiceSearch", "onNewIntent called, action: \${intent.action}, extras: \${intent.extras?.keySet()?.joinToString()}")
    if (intent.action == Intent.ACTION_SEARCH) {
      val query = intent.getStringExtra(SearchManager.QUERY)
      android.util.Log.d("VoiceSearch", "onNewIntent ACTION_SEARCH received, query: $query")
      if (!query.isNullOrBlank()) {
        val uri = android.net.Uri.parse("com.strmr.app:///search?q=\${android.net.Uri.encode(query)}")
        android.util.Log.d("VoiceSearch", "onNewIntent converting to deep link: $uri")
        super.onNewIntent(Intent(Intent.ACTION_VIEW, uri))
        return
      } else {
        android.util.Log.w("VoiceSearch", "onNewIntent ACTION_SEARCH but query was null/blank")
      }
    }
    android.util.Log.d("VoiceSearch", "onNewIntent passing through to super, action: \${intent.action}")
    super.onNewIntent(intent)
  }
`;

      // Insert before the last closing brace of the class
      const classMatch = content.match(/class MainActivity[^{]*\{/);
      if (classMatch) {
        const classStart = content.indexOf(classMatch[0]);
        let braceCount = 0;
        let classEnd = -1;

        for (let i = classStart; i < content.length; i++) {
          if (content[i] === '{') braceCount++;
          if (content[i] === '}') {
            braceCount--;
            if (braceCount === 0) {
              classEnd = i;
              break;
            }
          }
        }

        if (classEnd > 0) {
          content =
            content.slice(0, classEnd) + onNewIntentCode + content.slice(classEnd);
        }
      }

      fs.writeFileSync(mainActivityPath, content);
      console.log('\u2705 [VoiceSearch] Added voice search handling to MainActivity.kt');

      return config;
    },
  ]);
};

const withAndroidVoiceSearch = (config) => {
  if (process.env.EXPO_TV !== '1') {
    console.log('\u2139\ufe0f [VoiceSearch] Skipping voice search setup for non-TV build');
    return config;
  }

  config = withVoiceSearchManifest(config);
  config = withVoiceSearchResources(config);
  config = withVoiceSearchMainActivity(config);
  return config;
};

module.exports = withAndroidVoiceSearch;
