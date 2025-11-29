import { useState } from 'react';
import { Check, ArrowRight, ArrowLeft, Sparkles, Database as DatabaseIcon } from 'lucide-react';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import { useDatabaseTemplates, useCreateDatabase } from '@/features/database';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import type { DatabaseTemplate } from '@/features/database';

interface DatabaseWizardProps {
  onSuccess: () => void;
  onCancel: () => void;
}

export default function DatabaseWizard({ onSuccess, onCancel }: DatabaseWizardProps) {
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState({
    name: '',
    template: 'starter',
    display_name: '',
    description: '',
    shard_key: '',
  });

  const { data: templates, isLoading: templatesLoading } = useDatabaseTemplates();
  const createMutation = useCreateDatabase();

  const selectedTemplate = templates?.find((t) => t.name === formData.template);

  const handleNext = () => {
    if (step === 1 && !formData.name) {
      return; // Validation
    }
    if (step < 3) {
      setStep(step + 1);
    }
  };

  const handleBack = () => {
    if (step > 1) {
      setStep(step - 1);
    }
  };

  const handleSubmit = async () => {
    if (!formData.name) return;

    try {
      await createMutation.mutateAsync({
        name: formData.name,
        template: formData.template,
        display_name: formData.display_name || undefined,
        description: formData.description || undefined,
        shard_key: formData.shard_key || undefined,
      });
      onSuccess();
    } catch (error) {
      // Error handling is done by the mutation
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto">
      {/* Progress Steps */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {[1, 2, 3].map((s) => (
            <div key={s} className="flex items-center flex-1">
              <div className="flex flex-col items-center flex-1">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center font-semibold ${step >= s
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                    }`}
                >
                  {step > s ? <Check className="h-5 w-5" /> : s}
                </div>
                <div className="mt-2 text-xs text-center text-gray-600 dark:text-gray-400">
                  {s === 1 ? 'Basic Info' : s === 2 ? 'Template' : 'Review'}
                </div>
              </div>
              {s < 3 && (
                <div
                  className={`h-1 flex-1 mx-2 ${step > s ? 'bg-blue-600' : 'bg-gray-200 dark:bg-gray-700'
                    }`}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="min-h-[400px]">
        {step === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                Database Information
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                Let's start by giving your database a name and some basic details.
              </p>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Database Name <span className="text-red-500">*</span>
                </label>
                <Input
                  type="text"
                  placeholder="my-awesome-db"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  className="w-full"
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Use lowercase letters, numbers, and hyphens only
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Display Name (Optional)
                </label>
                <Input
                  type="text"
                  placeholder="My Awesome Database"
                  value={formData.display_name}
                  onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                  className="w-full"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Description (Optional)
                </label>
                <textarea
                  placeholder="Describe what this database will be used for..."
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:text-white"
                  rows={3}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Shard Key (Optional)
                </label>
                <Input
                  type="text"
                  placeholder="user_id (auto-detected if not provided)"
                  value={formData.shard_key}
                  onChange={(e) => setFormData({ ...formData, shard_key: e.target.value })}
                  className="w-full"
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  The column used to distribute data across shards
                </p>
              </div>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                Choose Template
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                Select a template that matches your needs. You can customize later.
              </p>
            </div>

            {templatesLoading ? (
              <div className="flex justify-center py-12">
                <LoadingSpinner />
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {templates?.map((template) => (
                  <TemplateCard
                    key={template.name}
                    template={template}
                    selected={formData.template === template.name}
                    onSelect={() => setFormData({ ...formData, template: template.name })}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {step === 3 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                Review & Create
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                Review your configuration and click Create to provision your database.
              </p>
            </div>

            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Name</div>
                  <div className="text-lg font-semibold text-gray-900 dark:text-white">
                    {formData.display_name || formData.name}
                  </div>
                </div>
                <div>
                  <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Template</div>
                  <div className="text-lg font-semibold text-gray-900 dark:text-white capitalize">
                    {formData.template}
                  </div>
                </div>
                {selectedTemplate && (
                  <>
                    <div>
                      <div className="text-sm font-medium text-gray-500 dark:text-gray-400">
                        Shards
                      </div>
                      <div className="text-lg font-semibold text-gray-900 dark:text-white">
                        {selectedTemplate.shard_count}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-gray-500 dark:text-gray-400">
                        Replicas per Shard
                      </div>
                      <div className="text-lg font-semibold text-gray-900 dark:text-white">
                        {selectedTemplate.replicas_per_shard}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-gray-500 dark:text-gray-400">
                        Estimated Cost
                      </div>
                      <div className="text-lg font-semibold text-gray-900 dark:text-white">
                        {selectedTemplate.estimated_cost}
                      </div>
                    </div>
                  </>
                )}
                {formData.shard_key && (
                  <div>
                    <div className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Shard Key
                    </div>
                    <div className="text-lg font-semibold text-gray-900 dark:text-white">
                      {formData.shard_key}
                    </div>
                  </div>
                )}
              </div>
              {formData.description && (
                <div>
                  <div className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">
                    Description
                  </div>
                  <div className="text-gray-900 dark:text-white">{formData.description}</div>
                </div>
              )}
            </div>

            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
              <div className="flex items-start">
                <Sparkles className="h-5 w-5 text-blue-600 dark:text-blue-400 mr-3 mt-0.5" />
                <div>
                  <div className="font-medium text-blue-900 dark:text-blue-100">
                    Zero-Touch Provisioning
                  </div>
                  <div className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                    Your database will be automatically provisioned with {selectedTemplate?.shard_count || 2}{' '}
                    shards and {selectedTemplate?.replicas_per_shard || 1} replica per shard. The
                    connection string will be available once provisioning completes (usually within 2
                    minutes).
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-between mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
        <div>
          {step > 1 && (
            <Button variant="secondary" onClick={handleBack}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back
            </Button>
          )}
        </div>
        <div className="flex gap-3">
          <Button variant="secondary" onClick={onCancel}>
            Cancel
          </Button>
          {step < 3 ? (
            <Button onClick={handleNext} disabled={step === 1 && !formData.name}>
              Next
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          ) : (
            <Button
              onClick={handleSubmit}
              disabled={createMutation.isPending || !formData.name}
            >
              {createMutation.isPending ? (
                <>
                  <LoadingSpinner className="mr-2" />
                  Creating...
                </>
              ) : (
                <>
                  <DatabaseIcon className="h-4 w-4 mr-2" />
                  Create Database
                </>
              )}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

interface TemplateCardProps {
  template: DatabaseTemplate;
  selected: boolean;
  onSelect: () => void;
}

function TemplateCard({ template, selected, onSelect }: TemplateCardProps) {
  return (
    <button
      onClick={onSelect}
      className={`p-6 rounded-lg border-2 text-left transition-all ${selected
        ? 'border-blue-600 bg-blue-50 dark:bg-blue-900/20 dark:border-blue-500'
        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
        }`}
    >
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
            {template.display_name}
          </h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            {template.description}
          </p>
        </div>
        {selected && (
          <div className="w-6 h-6 rounded-full bg-blue-600 flex items-center justify-center">
            <Check className="h-4 w-4 text-white" />
          </div>
        )}
      </div>
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-600 dark:text-gray-400">Shards:</span>
          <span className="font-medium text-gray-900 dark:text-white">{template.shard_count}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-600 dark:text-gray-400">Replicas:</span>
          <span className="font-medium text-gray-900 dark:text-white">
            {template.replicas_per_shard}
          </span>
        </div>
        <div className="flex justify-between pt-2 border-t border-gray-200 dark:border-gray-700">
          <span className="text-gray-600 dark:text-gray-400">Cost:</span>
          <span className="font-semibold text-blue-600 dark:text-blue-400">
            {template.estimated_cost}
          </span>
        </div>
      </div>
    </button>
  );
}

