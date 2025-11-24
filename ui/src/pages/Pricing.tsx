import { useState, useEffect } from 'react';
import { Check, X, Shield, Zap, Database } from 'lucide-react';
import { motion } from 'framer-motion';

interface PricingTier {
    name: string;
    price: string;
    description: string;
    features: string[];
    notIncluded: string[];
    buttonText: string;
    popular?: boolean;
    color: string;
    icon: any;
}

export default function Pricing() {
    const [currentPlan, setCurrentPlan] = useState<string>('Free');

    useEffect(() => {
        fetch('/api/v1/pricing')
            .then((res) => res.json())
            .then((data) => {
                // Normalize plan name to match UI (e.g. "free" -> "Free")
                const planName = data.Name.charAt(0).toUpperCase() + data.Name.slice(1).toLowerCase();
                setCurrentPlan(planName);
            })
            .catch((err) => console.error('Failed to fetch pricing:', err));
    }, []);

    const tiers: PricingTier[] = [
        {
            name: 'Free',
            price: '$0',
            description: 'Perfect for learning and small experiments.',
            features: [
                'Up to 2 Shards',
                '10 Requests per Second',
                'Eventual Consistency',
                'Community Support',
                'Basic Metrics',
            ],
            notIncluded: [
                'Strong Consistency',
                'Priority Support',
                'Custom Sharding Keys',
                'SLA Guarantee',
            ],
            buttonText: currentPlan === 'Free' ? 'Current Plan' : 'Downgrade',
            color: 'blue',
            icon: Database,
        },
        {
            name: 'Pro',
            price: '$49',
            description: 'For growing applications that need performance.',
            features: [
                'Up to 10 Shards',
                '100 Requests per Second',
                'Strong + Eventual Consistency',
                'Email Support',
                'Advanced Metrics',
                'Custom Sharding Keys',
            ],
            notIncluded: [
                'Unlimited Shards',
                '24/7 Priority Support',
                'Dedicated Infrastructure',
            ],
            buttonText: currentPlan === 'Pro' ? 'Current Plan' : 'Upgrade to Pro',
            popular: true,
            color: 'purple',
            icon: Zap,
        },
        {
            name: 'Enterprise',
            price: 'Custom',
            description: 'For large-scale mission-critical systems.',
            features: [
                'Unlimited Shards',
                'Unlimited RPS',
                'Strong + Eventual Consistency',
                '24/7 Priority Support',
                'Dedicated Infrastructure',
                'Custom SLA',
                'On-premise Deployment Option',
            ],
            notIncluded: [],
            buttonText: currentPlan === 'Enterprise' ? 'Current Plan' : 'Contact Sales',
            color: 'indigo',
            icon: Shield,
        },
    ];

    const handleUpgrade = (tierName: string) => {
        if (tierName === currentPlan) return;
        alert(`To upgrade to ${tierName}, please update your configuration file and restart the backend services.\n\nExample config:\n{\n  "pricing": {\n    "tier": "${tierName.toLowerCase()}"\n  }\n}`);
    };

    const container = {
        hidden: { opacity: 0 },
        show: {
            opacity: 1,
            transition: {
                staggerChildren: 0.1
            }
        }
    };

    const item = {
        hidden: { opacity: 0, y: 20 },
        show: { opacity: 1, y: 0 }
    };

    return (
        <div className="space-y-12">
            <div className="text-center max-w-3xl mx-auto pt-8">
                <motion.h1
                    initial={{ opacity: 0, y: -20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="text-4xl font-bold text-gray-900 dark:text-white mb-4"
                >
                    Simple, Transparent Pricing
                </motion.h1>
                <motion.p
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.2 }}
                    className="text-xl text-gray-600 dark:text-gray-400"
                >
                    Choose the perfect plan for your scaling needs. All plans include our core sharding technology.
                </motion.p>
            </div>

            <motion.div
                variants={container}
                initial="hidden"
                animate="show"
                className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-7xl mx-auto px-4"
            >
                {tiers.map((tier) => (
                    <motion.div
                        key={tier.name}
                        variants={item}
                        whileHover={{ y: -8 }}
                        className={`relative rounded-2xl border ${tier.popular
                            ? 'border-purple-500 shadow-xl z-10'
                            : 'border-gray-200 dark:border-gray-700 shadow-lg'
                            } bg-white dark:bg-gray-800 p-8 flex flex-col transition-all duration-300`}
                    >
                        {tier.popular && currentPlan !== tier.name && (
                            <div className="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                                <span className="bg-gradient-to-r from-purple-500 to-indigo-600 text-white px-4 py-1 rounded-full text-sm font-semibold uppercase tracking-wide shadow-lg">
                                    Most Popular
                                </span>
                            </div>
                        )}
                        {currentPlan === tier.name && (
                            <div className="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
                                <span className="bg-green-500 text-white px-4 py-1 rounded-full text-sm font-semibold uppercase tracking-wide shadow-lg">
                                    Current Plan
                                </span>
                            </div>
                        )}

                        <div className="mb-6">
                            <div className={`inline-flex p-3 rounded-xl bg-${tier.color}-50 dark:bg-${tier.color}-900/20 mb-4`}>
                                <tier.icon className={`h-8 w-8 text-${tier.color}-600 dark:text-${tier.color}-400`} />
                            </div>
                            <h3 className="text-2xl font-bold text-gray-900 dark:text-white">{tier.name}</h3>
                            <p className="mt-2 text-gray-500 dark:text-gray-400 min-h-[48px]">{tier.description}</p>
                        </div>

                        <div className="mb-8">
                            <span className="text-5xl font-bold text-gray-900 dark:text-white tracking-tight">{tier.price}</span>
                            {tier.price !== 'Custom' && <span className="text-gray-500 dark:text-gray-400 ml-2">/month</span>}
                        </div>

                        <ul className="space-y-4 mb-8 flex-1">
                            {tier.features.map((feature) => (
                                <li key={feature} className="flex items-start">
                                    <div className="flex-shrink-0 h-5 w-5 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mt-0.5 mr-3">
                                        <Check className="h-3 w-3 text-green-600 dark:text-green-400" />
                                    </div>
                                    <span className="text-gray-600 dark:text-gray-300 text-sm">{feature}</span>
                                </li>
                            ))}
                            {tier.notIncluded.map((feature) => (
                                <li key={feature} className="flex items-start opacity-50">
                                    <div className="flex-shrink-0 h-5 w-5 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mt-0.5 mr-3">
                                        <X className="h-3 w-3 text-gray-400" />
                                    </div>
                                    <span className="text-gray-500 dark:text-gray-400 text-sm">{feature}</span>
                                </li>
                            ))}
                        </ul>

                        <button
                            onClick={() => handleUpgrade(tier.name)}
                            disabled={currentPlan === tier.name}
                            className={`w-full py-4 px-6 rounded-xl font-bold transition-all duration-200 transform active:scale-95 ${currentPlan === tier.name
                                ? 'bg-gray-100 text-gray-400 cursor-not-allowed dark:bg-gray-700 dark:text-gray-500'
                                : tier.popular
                                    ? 'bg-gradient-to-r from-purple-600 to-indigo-600 text-white hover:from-purple-700 hover:to-indigo-700 shadow-lg hover:shadow-purple-500/30'
                                    : 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 shadow-md'
                                }`}
                        >
                            {tier.buttonText}
                        </button>
                    </motion.div>
                ))}
            </motion.div>

            <div className="mt-12 text-center pb-12">
                <p className="text-gray-500 dark:text-gray-400">
                    Need a custom solution? <a href="#" className="text-primary-600 hover:underline font-medium">Contact our sales team</a>
                </p>
            </div>
        </div>
    );
}
