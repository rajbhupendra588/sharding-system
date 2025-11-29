import { useState } from 'react';
import { Copy, Check, Database } from 'lucide-react';
import Modal from '@/components/ui/Modal';
import { toast } from 'react-hot-toast';

interface ConnectAppModalProps {
    isOpen: boolean;
    onClose: () => void;
    appName: string;
    databaseName: string;
    username?: string;
    password?: string;
}

type Language = 'go' | 'java' | 'node' | 'python' | 'cli';

export default function ConnectAppModal({
    isOpen,
    onClose,
    appName,
    databaseName,
    username = 'app_user',
    password = '***************',
}: ConnectAppModalProps) {
    const [activeTab, setActiveTab] = useState<Language>('go');
    const [copied, setCopied] = useState(false);

    const routerHost = window.location.hostname;
    const routerPort = '6000'; // Default Router Port

    const connectionDetails = {
        host: routerHost,
        port: routerPort,
        database: databaseName,
        username: username,
        password: password,
    };

    const snippets: Record<Language, string> = {
        cli: `mysql -h ${routerHost} -P ${routerPort} -u ${username} -p${password} ${databaseName}`,

        go: `import (
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
)

func main() {
  dsn := "${username}:${password}@tcp(${routerHost}:${routerPort})/${databaseName}?tls=true"
  db, err := sql.Open("mysql", dsn)
  if err != nil {
    panic(err)
  }
  defer db.Close()
}`,

        java: `// build.gradle
implementation 'mysql:mysql-connector-java:8.0.33'

// Application.java
String url = "jdbc:mysql://${routerHost}:${routerPort}/${databaseName}?sslMode=VERIFY_IDENTITY";
Properties props = new Properties();
props.setProperty("user", "${username}");
props.setProperty("password", "${password}");

Connection conn = DriverManager.getConnection(url, props);`,

        node: `const mysql = require('mysql2/promise');

const connection = await mysql.createConnection({
  host: '${routerHost}',
  port: ${routerPort},
  user: '${username}',
  password: '${password}',
  database: '${databaseName}',
  ssl: {
    rejectUnauthorized: true
  }
});`,

        python: `import mysql.connector

connection = mysql.connector.connect(
  host="${routerHost}",
  port=${routerPort},
  user="${username}",
  password="${password}",
  database="${databaseName}",
  ssl_ca="/etc/ssl/certs/ca-certificates.crt"
)`
    };

    const handleCopy = () => {
        navigator.clipboard.writeText(snippets[activeTab]);
        setCopied(true);
        toast.success('Snippet copied to clipboard');
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title={`Connect to ${appName}`}
            size="lg"
        >
            <div className="space-y-6">
                {/* Connection Details Cards */}
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                    <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                        <div className="text-xs text-gray-500 dark:text-gray-400 uppercase font-semibold mb-1">Host</div>
                        <div className="font-mono text-sm text-gray-900 dark:text-white truncate" title={connectionDetails.host}>
                            {connectionDetails.host}
                        </div>
                    </div>
                    <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                        <div className="text-xs text-gray-500 dark:text-gray-400 uppercase font-semibold mb-1">Port</div>
                        <div className="font-mono text-sm text-gray-900 dark:text-white">
                            {connectionDetails.port}
                        </div>
                    </div>
                    <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                        <div className="text-xs text-gray-500 dark:text-gray-400 uppercase font-semibold mb-1">User</div>
                        <div className="font-mono text-sm text-gray-900 dark:text-white truncate" title={connectionDetails.username}>
                            {connectionDetails.username}
                        </div>
                    </div>
                    <div className="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                        <div className="text-xs text-gray-500 dark:text-gray-400 uppercase font-semibold mb-1">Database</div>
                        <div className="font-mono text-sm text-gray-900 dark:text-white truncate" title={connectionDetails.database}>
                            {connectionDetails.database}
                        </div>
                    </div>
                </div>

                {/* Language Tabs */}
                <div>
                    <div className="flex items-center gap-2 border-b border-gray-200 dark:border-gray-700 mb-0">
                        {(['go', 'java', 'node', 'python', 'cli'] as Language[]).map((lang) => (
                            <button
                                key={lang}
                                onClick={() => setActiveTab(lang)}
                                className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === lang
                                    ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                                    : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                                    }`}
                            >
                                {lang === 'cli' ? 'CLI' : lang.charAt(0).toUpperCase() + lang.slice(1)}
                            </button>
                        ))}
                    </div>

                    {/* Code Snippet */}
                    <div className="relative mt-4">
                        <div className="absolute top-3 right-3">
                            <button
                                onClick={handleCopy}
                                className="p-2 text-gray-400 hover:text-white bg-gray-800 hover:bg-gray-700 rounded-md transition-colors"
                                title="Copy code"
                            >
                                {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4" />}
                            </button>
                        </div>
                        <pre className="p-4 bg-gray-900 text-gray-100 rounded-lg overflow-x-auto text-sm font-mono leading-relaxed">
                            <code>{snippets[activeTab]}</code>
                        </pre>
                    </div>
                </div>

                <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg flex items-start gap-3">
                    <Database className="h-5 w-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
                    <div className="text-sm text-blue-800 dark:text-blue-200">
                        <p className="font-medium mb-1">About Credentials</p>
                        <p>
                            Use these credentials to connect your application to the Sharding Router.
                            The Router will automatically route your queries to the correct shard based on your sharding key.
                        </p>
                    </div>
                </div>
            </div>
        </Modal>
    );
}
