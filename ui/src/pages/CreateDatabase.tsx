import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Database,
  Server,
  Layers,
  Zap,
  Shield,
  Check,
  ChevronRight,
  ChevronLeft,
  Sparkles,
  Copy,
  Terminal,
  Code,
  Box,
  ArrowRight
} from 'lucide-react';

// Template configurations
const databaseTemplates = [
  {
    id: 'starter',
    name: 'Starter',
    description: 'Perfect for development and small applications',
    shardCount: 2,
    cpu: '250m',
    memory: '512Mi',
    storage: '5Gi',
    icon: Zap,
    color: 'from-emerald-500 to-teal-600',
    recommended: false,
  },
  {
    id: 'production',
    name: 'Production',
    description: 'Balanced configuration for production workloads',
    shardCount: 4,
    cpu: '1000m',
    memory: '2Gi',
    storage: '50Gi',
    icon: Server,
    color: 'from-blue-500 to-indigo-600',
    recommended: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    description: 'High-performance for large scale applications',
    shardCount: 8,
    cpu: '2000m',
    memory: '4Gi',
    storage: '100Gi',
    icon: Shield,
    color: 'from-purple-500 to-pink-600',
    recommended: false,
  },
];

const schemaTemplates = [
  {
    id: 'users',
    name: 'Users',
    description: 'Standard users table with authentication',
    tables: ['users'],
    icon: '👤',
  },
  {
    id: 'orders',
    name: 'E-commerce',
    description: 'Orders and order items',
    tables: ['orders', 'order_items'],
    icon: '🛒',
  },
  {
    id: 'products',
    name: 'Products',
    description: 'Product catalog with inventory',
    tables: ['products', 'categories', 'inventory'],
    icon: '📦',
  },
  {
    id: 'analytics',
    name: 'Analytics',
    description: 'Events and sessions tracking',
    tables: ['events', 'sessions'],
    icon: '📊',
  },
  {
    id: 'custom',
    name: 'Custom',
    description: 'Start with empty database',
    tables: [],
    icon: '✨',
  },
];

const steps = [
  { id: 1, name: 'Template', description: 'Choose configuration' },
  { id: 2, name: 'Details', description: 'Database settings' },
  { id: 3, name: 'Schema', description: 'Initial tables' },
  { id: 4, name: 'Review', description: 'Confirm & create' },
];

export default function CreateDatabase() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [isCreating, setIsCreating] = useState(false);
  const [createdDatabase, setCreatedDatabase] = useState<any>(null);

  // Form state
  const [selectedTemplate, setSelectedTemplate] = useState('production');
  const [databaseName, setDatabaseName] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [description, setDescription] = useState('');
  const [shardKey, setShardKey] = useState('user_id');
  const [shardKeyType, setShardKeyType] = useState('uuid');
  const [strategy, setStrategy] = useState('hash');
  const [customShardCount] = useState<number | null>(null);
  const [selectedSchema, setSelectedSchema] = useState('users');
  const [customSQL, setCustomSQL] = useState('');

  const template = databaseTemplates.find(t => t.id === selectedTemplate)!;
  const schemaTemplate = schemaTemplates.find(s => s.id === selectedSchema);

  const handleCreate = async () => {
    setIsCreating(true);

    try {
      const response = await fetch('/api/v1/databases', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: databaseName,
          display_name: displayName || databaseName,
          description,
          template: selectedTemplate,
          shard_count: customShardCount || template.shardCount,
          shard_key: shardKey,
          shard_key_type: shardKeyType,
          strategy,
          schema_template: selectedSchema !== 'custom' ? selectedSchema : undefined,
          schema: selectedSchema === 'custom' ? customSQL : undefined,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to create database');
      }

      const data = await response.json();
      setCreatedDatabase(data);
      setCurrentStep(5); // Success step
    } catch (error) {
      console.error('Error creating database:', error);
    } finally {
      setIsCreating(false);
    }
  };

  const canProceed = () => {
    switch (currentStep) {
      case 1:
        return selectedTemplate !== '';
      case 2:
        return databaseName.length >= 3 && /^[a-z][a-z0-9_]*$/.test(databaseName);
      case 3:
        return selectedSchema !== '';
      case 4:
        return true;
      default:
        return false;
    }
  };

  return (
    <div className="min-h-screen bg-[#0a0a0f] text-white">
      {/* Animated background */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-900/20 via-transparent to-purple-900/20" />
        <motion.div
          className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl"
          animate={{
            scale: [1, 1.2, 1],
            opacity: [0.3, 0.5, 0.3]
          }}
          transition={{ duration: 8, repeat: Infinity }}
        />
        <motion.div
          className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl"
          animate={{
            scale: [1.2, 1, 1.2],
            opacity: [0.3, 0.5, 0.3]
          }}
          transition={{ duration: 8, repeat: Infinity }}
        />
      </div>

      <div className="relative max-w-5xl mx-auto px-6 py-12">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-blue-500/20 to-purple-500/20 border border-blue-500/30 mb-6">
            <Sparkles className="w-4 h-4 text-blue-400" />
            <span className="text-sm text-blue-300">Zero-Config Database Creation</span>
          </div>
          <h1 className="text-4xl font-bold mb-3 bg-gradient-to-r from-white via-blue-100 to-purple-200 bg-clip-text text-transparent">
            Create Sharded Database
          </h1>
          <p className="text-gray-400 max-w-xl mx-auto">
            Launch a production-ready sharded PostgreSQL cluster in minutes. No infrastructure experience required.
          </p>
        </motion.div>

        {/* Progress Steps */}
        {currentStep < 5 && (
          <div className="mb-12">
            <div className="flex items-center justify-center gap-4">
              {steps.map((step, index) => (
                <React.Fragment key={step.id}>
                  <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: index * 0.1 }}
                    className={`flex items-center gap-3 ${currentStep === step.id
                        ? 'text-white'
                        : currentStep > step.id
                          ? 'text-green-400'
                          : 'text-gray-500'
                      }`}
                  >
                    <div className={`
                      w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium
                      ${currentStep === step.id
                        ? 'bg-gradient-to-r from-blue-500 to-purple-500 shadow-lg shadow-blue-500/30'
                        : currentStep > step.id
                          ? 'bg-green-500/20 border border-green-500/50'
                          : 'bg-gray-800 border border-gray-700'}
                    `}>
                      {currentStep > step.id ? <Check className="w-5 h-5" /> : step.id}
                    </div>
                    <div className="hidden sm:block">
                      <div className="text-sm font-medium">{step.name}</div>
                      <div className="text-xs text-gray-500">{step.description}</div>
                    </div>
                  </motion.div>
                  {index < steps.length - 1 && (
                    <div className={`w-12 h-0.5 ${currentStep > step.id ? 'bg-green-500' : 'bg-gray-700'}`} />
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        )}

        {/* Step Content */}
        <AnimatePresence mode="wait">
          {/* Step 1: Template Selection */}
          {currentStep === 1 && (
            <motion.div
              key="step1"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="space-y-6"
            >
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {databaseTemplates.map((t) => {
                  const Icon = t.icon;
                  return (
                    <motion.button
                      key={t.id}
                      whileHover={{ scale: 1.02 }}
                      whileTap={{ scale: 0.98 }}
                      onClick={() => setSelectedTemplate(t.id)}
                      className={`
                        relative p-6 rounded-2xl text-left transition-all
                        ${selectedTemplate === t.id
                          ? 'bg-gradient-to-br ' + t.color + ' shadow-xl'
                          : 'bg-gray-800/50 border border-gray-700 hover:border-gray-600'}
                      `}
                    >
                      {t.recommended && (
                        <div className="absolute -top-3 left-1/2 -translate-x-1/2 px-3 py-1 rounded-full bg-yellow-500 text-black text-xs font-medium">
                          Recommended
                        </div>
                      )}
                      <Icon className={`w-8 h-8 mb-4 ${selectedTemplate === t.id ? 'text-white' : 'text-gray-400'}`} />
                      <h3 className="text-lg font-semibold mb-2">{t.name}</h3>
                      <p className="text-sm text-gray-300 mb-4">{t.description}</p>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between text-gray-300">
                          <span>Shards</span>
                          <span className="font-mono">{t.shardCount}</span>
                        </div>
                        <div className="flex justify-between text-gray-300">
                          <span>CPU</span>
                          <span className="font-mono">{t.cpu}</span>
                        </div>
                        <div className="flex justify-between text-gray-300">
                          <span>Memory</span>
                          <span className="font-mono">{t.memory}</span>
                        </div>
                        <div className="flex justify-between text-gray-300">
                          <span>Storage</span>
                          <span className="font-mono">{t.storage}</span>
                        </div>
                      </div>
                    </motion.button>
                  );
                })}
              </div>
            </motion.div>
          )}

          {/* Step 2: Database Details */}
          {currentStep === 2 && (
            <motion.div
              key="step2"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="max-w-2xl mx-auto space-y-6"
            >
              <div className="bg-gray-800/50 rounded-2xl border border-gray-700 p-6 space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    Database Name *
                  </label>
                  <input
                    type="text"
                    value={databaseName}
                    onChange={(e) => setDatabaseName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
                    placeholder="my_database"
                    className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all font-mono"
                  />
                  <p className="mt-2 text-xs text-gray-500">
                    Lowercase letters, numbers, and underscores only. Must start with a letter.
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    Display Name
                  </label>
                  <input
                    type="text"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="My Database"
                    className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    Description
                  </label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="What is this database for?"
                    rows={3}
                    className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all resize-none"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                      Shard Key
                    </label>
                    <input
                      type="text"
                      value={shardKey}
                      onChange={(e) => setShardKey(e.target.value)}
                      placeholder="user_id"
                      className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all font-mono"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                      Key Type
                    </label>
                    <select
                      value={shardKeyType}
                      onChange={(e) => setShardKeyType(e.target.value)}
                      className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
                    >
                      <option value="uuid">UUID</option>
                      <option value="integer">Integer</option>
                      <option value="string">String</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    Sharding Strategy
                  </label>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      onClick={() => setStrategy('hash')}
                      className={`p-4 rounded-xl border transition-all ${strategy === 'hash'
                          ? 'border-blue-500 bg-blue-500/10'
                          : 'border-gray-600 hover:border-gray-500'
                        }`}
                    >
                      <Layers className="w-6 h-6 mb-2 text-blue-400" />
                      <div className="font-medium">Hash</div>
                      <div className="text-xs text-gray-400">Even distribution</div>
                    </button>
                    <button
                      onClick={() => setStrategy('range')}
                      className={`p-4 rounded-xl border transition-all ${strategy === 'range'
                          ? 'border-purple-500 bg-purple-500/10'
                          : 'border-gray-600 hover:border-gray-500'
                        }`}
                    >
                      <Box className="w-6 h-6 mb-2 text-purple-400" />
                      <div className="font-medium">Range</div>
                      <div className="text-xs text-gray-400">Sequential access</div>
                    </button>
                  </div>
                </div>
              </div>
            </motion.div>
          )}

          {/* Step 3: Schema Selection */}
          {currentStep === 3 && (
            <motion.div
              key="step3"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="space-y-6"
            >
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                {schemaTemplates.map((s) => (
                  <motion.button
                    key={s.id}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={() => setSelectedSchema(s.id)}
                    className={`
                      p-4 rounded-xl text-center transition-all
                      ${selectedSchema === s.id
                        ? 'bg-gradient-to-br from-blue-500 to-purple-500 shadow-lg'
                        : 'bg-gray-800/50 border border-gray-700 hover:border-gray-600'}
                    `}
                  >
                    <div className="text-3xl mb-2">{s.icon}</div>
                    <div className="font-medium text-sm">{s.name}</div>
                    <div className="text-xs text-gray-400 mt-1">{s.tables.length} tables</div>
                  </motion.button>
                ))}
              </div>

              {selectedSchema && schemaTemplate && (
                <div className="bg-gray-800/50 rounded-2xl border border-gray-700 p-6">
                  <h3 className="font-semibold mb-2">{schemaTemplate.name}</h3>
                  <p className="text-sm text-gray-400 mb-4">{schemaTemplate.description}</p>
                  {schemaTemplate.tables.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                      {schemaTemplate.tables.map((table) => (
                        <span key={table} className="px-3 py-1 rounded-full bg-gray-700 text-sm font-mono">
                          {table}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {selectedSchema === 'custom' && (
                <div className="bg-gray-800/50 rounded-2xl border border-gray-700 p-6">
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    Custom SQL Schema
                  </label>
                  <textarea
                    value={customSQL}
                    onChange={(e) => setCustomSQL(e.target.value)}
                    placeholder="CREATE TABLE users (...);"
                    rows={10}
                    className="w-full px-4 py-3 bg-gray-900 border border-gray-600 rounded-xl focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all resize-none font-mono text-sm"
                  />
                </div>
              )}
            </motion.div>
          )}

          {/* Step 4: Review */}
          {currentStep === 4 && (
            <motion.div
              key="step4"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="max-w-2xl mx-auto"
            >
              <div className="bg-gradient-to-br from-gray-800/80 to-gray-900/80 rounded-2xl border border-gray-700 overflow-hidden">
                <div className="p-6 border-b border-gray-700">
                  <div className="flex items-center gap-4">
                    <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${template.color} flex items-center justify-center`}>
                      <Database className="w-6 h-6" />
                    </div>
                    <div>
                      <h2 className="text-xl font-bold">{displayName || databaseName}</h2>
                      <p className="text-sm text-gray-400">{description || 'No description'}</p>
                    </div>
                  </div>
                </div>

                <div className="p-6 space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Template</div>
                      <div className="font-medium">{template.name}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Shards</div>
                      <div className="font-medium">{customShardCount || template.shardCount}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Strategy</div>
                      <div className="font-medium capitalize">{strategy}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Shard Key</div>
                      <div className="font-medium font-mono">{shardKey}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Schema</div>
                      <div className="font-medium">{schemaTemplate?.name}</div>
                    </div>
                    <div className="p-4 bg-gray-800/50 rounded-xl">
                      <div className="text-xs text-gray-500 uppercase mb-1">Resources</div>
                      <div className="font-medium">{template.cpu} / {template.memory}</div>
                    </div>
                  </div>

                  <div className="p-4 bg-blue-500/10 border border-blue-500/30 rounded-xl">
                    <div className="flex items-center gap-2 text-blue-400 mb-2">
                      <Zap className="w-4 h-4" />
                      <span className="text-sm font-medium">What will be created</span>
                    </div>
                    <ul className="text-sm text-gray-300 space-y-1">
                      <li>• {customShardCount || template.shardCount} PostgreSQL instances (shards)</li>
                      <li>• {customShardCount || template.shardCount} Persistent volumes ({template.storage} each)</li>
                      <li>• Automatic load balancing and routing</li>
                      <li>• Health monitoring and auto-recovery</li>
                    </ul>
                  </div>
                </div>
              </div>
            </motion.div>
          )}

          {/* Step 5: Success */}
          {currentStep === 5 && createdDatabase && (
            <motion.div
              key="step5"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="max-w-2xl mx-auto text-center"
            >
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ type: 'spring', delay: 0.2 }}
                className="w-20 h-20 mx-auto mb-6 rounded-full bg-gradient-to-br from-green-400 to-emerald-600 flex items-center justify-center"
              >
                <Check className="w-10 h-10 text-white" />
              </motion.div>

              <h2 className="text-2xl font-bold mb-2">Database Created!</h2>
              <p className="text-gray-400 mb-8">
                Your sharded database is being provisioned. It will be ready in a few minutes.
              </p>

              <div className="bg-gray-800/50 rounded-2xl border border-gray-700 p-6 text-left mb-6">
                <h3 className="font-semibold mb-4 flex items-center gap-2">
                  <Terminal className="w-5 h-5 text-blue-400" />
                  Connection String
                </h3>
                <div className="relative">
                  <code className="block p-4 bg-gray-900 rounded-xl text-sm font-mono text-green-400 overflow-x-auto">
                    {createdDatabase.connection_string || `postgresql://sharding_admin@sharding-proxy:6432/${databaseName}`}
                  </code>
                  <button
                    onClick={() => navigator.clipboard.writeText(createdDatabase.connection_string)}
                    className="absolute top-3 right-3 p-2 rounded-lg bg-gray-700 hover:bg-gray-600 transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
              </div>

              <div className="bg-gray-800/50 rounded-2xl border border-gray-700 p-6 text-left mb-8">
                <h3 className="font-semibold mb-4 flex items-center gap-2">
                  <Code className="w-5 h-5 text-purple-400" />
                  Quick Start
                </h3>
                <pre className="p-4 bg-gray-900 rounded-xl text-sm font-mono overflow-x-auto">
                  {`// Node.js
const { Client } = require('pg');
const client = new Client('${createdDatabase.connection_string || `postgresql://...`}');
await client.connect();

// Java
String url = "${createdDatabase.connection_string || `jdbc:postgresql://...`}";
Connection conn = DriverManager.getConnection(url);

// Go
db, _ := sql.Open("postgres", "${createdDatabase.connection_string || `postgresql://...`}")`}
                </pre>
              </div>

              <div className="flex gap-4 justify-center">
                <button
                  onClick={() => navigate('/databases')}
                  className="px-6 py-3 rounded-xl bg-gray-700 hover:bg-gray-600 transition-colors"
                >
                  View All Databases
                </button>
                <button
                  onClick={() => navigate(`/databases/${createdDatabase.name}`)}
                  className="px-6 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 transition-colors flex items-center gap-2"
                >
                  Open Dashboard
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Navigation Buttons */}
        {currentStep < 5 && (
          <div className="flex justify-between mt-8">
            <button
              onClick={() => setCurrentStep(Math.max(1, currentStep - 1))}
              disabled={currentStep === 1}
              className="px-6 py-3 rounded-xl bg-gray-700 hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <ChevronLeft className="w-4 h-4" />
              Back
            </button>

            {currentStep < 4 ? (
              <button
                onClick={() => setCurrentStep(currentStep + 1)}
                disabled={!canProceed()}
                className="px-6 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                Continue
                <ChevronRight className="w-4 h-4" />
              </button>
            ) : (
              <button
                onClick={handleCreate}
                disabled={isCreating}
                className="px-8 py-3 rounded-xl bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 disabled:opacity-50 transition-colors flex items-center gap-2"
              >
                {isCreating ? (
                  <>
                    <motion.div
                      animate={{ rotate: 360 }}
                      transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                      className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full"
                    />
                    Creating...
                  </>
                ) : (
                  <>
                    <Sparkles className="w-4 h-4" />
                    Create Database
                  </>
                )}
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}





