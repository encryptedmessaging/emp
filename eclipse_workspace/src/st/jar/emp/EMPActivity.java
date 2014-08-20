package st.jar.emp;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

import android.support.v7.app.ActionBarActivity;
import android.annotation.SuppressLint;
import android.content.res.AssetManager;
import android.os.Bundle;
import android.view.Menu;
import android.view.MenuItem;
import android.view.Window;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;


public class EMPActivity extends ActionBarActivity {
	
	private WebView myWebView;
	private Process empProcess;
	
    @SuppressLint("SetJavaScriptEnabled")
	@Override
    protected void onCreate(Bundle savedInstanceState) {
    	Process p;
    	
    	//Remove title bar
        this.requestWindowFeature(Window.FEATURE_NO_TITLE);
        
        super.onCreate(savedInstanceState);
        
        setContentView(R.layout.activity_emp);
        
        System.out.println("Creating Application...");
        
        // Start Web View
        myWebView = (WebView) findViewById(R.id.webview);
        
        myWebView.loadUrl("http://127.0.0.1:8080");
        
        WebSettings webSettings = myWebView.getSettings();
        webSettings.setJavaScriptEnabled(true);
        myWebView.setWebViewClient(new WebViewClient());
        webSettings.setUseWideViewPort(true);
        webSettings.setLoadWithOverviewMode(true);
        
        empProcess = null;
        
        // Populate Local Storage
        try {
			p = Runtime.getRuntime().exec(new String[]{"/system/bin/rm", "-r", "-f", getFilesDir() + "/config/client/"});
			p.waitFor();
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
			System.exit(-1);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
			System.exit(-1);
		}
        
        copyAssetFolder(getApplicationContext().getAssets(), "config", getFilesDir() + "/config/");
        
        
    }

    @Override
    protected void onStart() {
    	super.onStart();
    	
    	System.out.println("Starting...");
    	if (empProcess != null) {
    		empProcess.destroy();
    		try {
				empProcess.waitFor();
			} catch (InterruptedException e) {
				// TODO Auto-generated catch block
				e.printStackTrace();
				System.exit(-1);
			}
    		empProcess = null;
    	}
    	
    	// Start new process
		try {
			empProcess = Runtime.getRuntime().exec(new String[]{getApplicationInfo().nativeLibraryDir + "/libemp.so", getFilesDir() + "/config/"});
		} catch (IOException e) {
			e.printStackTrace();
			System.exit(-1);
		}
		
    }
	
    @Override
    protected void onStop() {
    	super.onStop();
    	
    	System.out.println("Stopping...");
    	
    	// Stop EMP Process
    	if (empProcess != null) {
    		empProcess.destroy();
    		empProcess = null;
    	}
    }
    
    @Override
    public void onBackPressed() {
        myWebView.loadUrl("javascript:BackButton()");
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();
        if (id == R.id.action_settings) {
            return true;
        }
        return super.onOptionsItemSelected(item);
    }
    
    // Asset Copy Functions
    
    private static boolean copyAssetFolder(AssetManager assetManager,
            String fromAssetPath, String toPath) {
        try {
            String[] files = assetManager.list(fromAssetPath);
            new File(toPath).mkdirs();
            boolean res = true;
            for (String file : files)
                if (file.contains("."))
                    res &= copyAsset(assetManager, 
                            fromAssetPath + "/" + file,
                            toPath + "/" + file);
                else 
                    res &= copyAssetFolder(assetManager, 
                            fromAssetPath + "/" + file,
                            toPath + "/" + file);
            return res;
        } catch (Exception e) {
            e.printStackTrace();
            return false;
        }
    }

    private static boolean copyAsset(AssetManager assetManager,
            String fromAssetPath, String toPath) {
        InputStream in = null;
        OutputStream out = null;
        try {
          in = assetManager.open(fromAssetPath);
          new File(toPath).createNewFile();
          out = new FileOutputStream(toPath);
          copyFile(in, out);
          in.close();
          in = null;
          out.flush();
          out.close();
          out = null;
          return true;
        } catch(Exception e) {
            e.printStackTrace();
            return false;
        }
    }

    private static void copyFile(InputStream in, OutputStream out) throws IOException {
        byte[] buffer = new byte[1024];
        int read;
        while((read = in.read(buffer)) != -1){
          out.write(buffer, 0, read);
        }
    }
}
