# Feature Flags Architecture

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application                              │
│                                                                  │
│  ┌────────────────┐      ┌──────────────────┐                  │
│  │   Handlers     │      │    Usecases      │                  │
│  │                │      │                  │                  │
│  │  - GetUsers()  │      │  - CreateUser()  │                  │
│  │  - Dashboard() │      │  - SendEmail()   │                  │
│  │  - API()       │      │  - Process()     │                  │
│  └────────┬───────┘      └────────┬─────────┘                  │
│           │                       │                             │
│           │  featureflag.IsEnabled(ctx, "flag", default)       │
│           │                       │                             │
│           └───────────┬───────────┘                             │
│                       │                                         │
│                       ▼                                         │
│           ┌───────────────────────┐                            │
│           │  Feature Flag Client  │                            │
│           │                       │                            │
│           │  - IsEnabled()        │                            │
│           │  - GetStringVariation()│                           │
│           │  - GetIntVariation()  │                            │
│           │  - GetJSONVariation() │                            │
│           └───────────┬───────────┘                            │
│                       │                                         │
└───────────────────────┼─────────────────────────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────┐
        │   GoFeatureFlag Library       │
        │                               │
        │  - Flag Evaluation            │
        │  - Targeting Rules            │
        │  - Percentage Rollout         │
        └───────────┬───────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
┌───────────────┐       ┌──────────────┐
│ File Retriever│       │Redis Retriever│
│               │       │              │
│ flags.yaml    │       │ Redis Key    │
└───────────────┘       └──────────────┘
```

## Data Flow

```
1. Request arrives
   │
   ▼
2. Middleware injects Feature Flag Client into context
   │
   ▼
3. Handler/Usecase calls featureflag.IsEnabled(ctx, "flag", default)
   │
   ▼
4. Helper extracts Client and User from context
   │
   ▼
5. Client evaluates flag with user attributes
   │
   ├─► Check targeting rules (email, plan, country, etc.)
   ├─► Check percentage rollout
   ├─► Check time-based scheduling
   └─► Return variation
   │
   ▼
6. Handler/Usecase executes appropriate code path
   │
   ▼
7. Response sent to client
```

## Component Interaction

```
┌──────────────┐
│   Request    │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│   Middleware                 │
│   - Inject FF Client         │
│   - Extract User Info        │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│   Handler                    │
│                              │
│   user := ff.NewUser(id)     │
│   ctx = ff.WithUser(ctx, u)  │
│                              │
│   if ff.IsEnabled(ctx, ...)  │
│     → New Feature            │
│   else                       │
│     → Old Feature            │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│   Feature Flag Client        │
│                              │
│   1. Get user from context   │
│   2. Evaluate flag           │
│   3. Apply targeting rules   │
│   4. Return result           │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│   Configuration Source       │
│                              │
│   File: flags.yaml           │
│   Redis: feature_flags key   │
└──────────────────────────────┘
```

## Flag Evaluation Process

```
┌─────────────────────────────────────────────────────────────┐
│                    Flag Evaluation                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  Get Flag Definition  │
                └───────────┬───────────┘
                            │
                            ▼
                ┌───────────────────────┐
                │  Check if Disabled    │
                └───────────┬───────────┘
                            │
                    ┌───────┴───────┐
                    │               │
                    ▼               ▼
            ┌──────────┐      ┌──────────┐
            │   Yes    │      │    No    │
            │  Return  │      │ Continue │
            │  Default │      └────┬─────┘
            └──────────┘           │
                                   ▼
                        ┌──────────────────┐
                        │ Check Scheduling │
                        └────────┬─────────┘
                                 │
                        ┌────────┴────────┐
                        │                 │
                        ▼                 ▼
                  ┌──────────┐      ┌──────────┐
                  │ In Range │      │Out Range │
                  │ Continue │      │  Return  │
                  └────┬─────┘      │  Default │
                       │            └──────────┘
                       ▼
            ┌──────────────────────┐
            │  Check Targeting     │
            │  Rules               │
            └──────────┬───────────┘
                       │
            ┌──────────┴──────────┐
            │                     │
            ▼                     ▼
      ┌──────────┐          ┌──────────┐
      │  Match   │          │No Match  │
      │  Return  │          │ Continue │
      │ Variation│          └────┬─────┘
      └──────────┘               │
                                 ▼
                      ┌──────────────────┐
                      │ Apply Percentage │
                      │    Rollout       │
                      └────────┬─────────┘
                               │
                               ▼
                      ┌──────────────────┐
                      │ Return Variation │
                      └──────────────────┘
```

## User Context Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    User Context                             │
└─────────────────────────────────────────────────────────────┘

1. Create User
   ┌────────────────────────────────────┐
   │ user := ff.NewUser("user-123")     │
   │   .WithEmail("user@example.com")   │
   │   .WithCustom("plan", "premium")   │
   │   .WithCustom("beta", true)        │
   └────────────────────────────────────┘
                    │
                    ▼
2. Add to Context
   ┌────────────────────────────────────┐
   │ ctx = ff.WithUser(ctx, user)       │
   └────────────────────────────────────┘
                    │
                    ▼
3. Evaluate Flags
   ┌────────────────────────────────────┐
   │ ff.IsEnabled(ctx, "flag", default) │
   └────────────────────────────────────┘
                    │
                    ▼
4. Extract User & Evaluate
   ┌────────────────────────────────────┐
   │ user, ok := ff.GetUser(ctx)        │
   │ if !ok {                           │
   │   user = ff.NewAnonymousUser()     │
   │ }                                  │
   │                                    │
   │ evalCtx := user.ToEvaluationContext()│
   │ result := ffclient.BoolVariation(  │
   │   flagKey, evalCtx, default)       │
   └────────────────────────────────────┘
```

## Configuration Refresh

```
┌─────────────────────────────────────────────────────────────┐
│              Configuration Refresh Cycle                    │
└─────────────────────────────────────────────────────────────┘

     ┌──────────────────┐
     │  Application     │
     │  Starts          │
     └────────┬─────────┘
              │
              ▼
     ┌──────────────────┐
     │  Load Initial    │
     │  Configuration   │
     └────────┬─────────┘
              │
              ▼
     ┌──────────────────┐
     │  Start Polling   │
     │  (every 60s)     │
     └────────┬─────────┘
              │
              ▼
        ┌─────────┐
        │  Wait   │
        └────┬────┘
             │
             ▼
     ┌──────────────────┐
     │  Fetch Config    │
     │  from Source     │
     └────────┬─────────┘
              │
              ▼
     ┌──────────────────┐
     │  Compare with    │
     │  Current Config  │
     └────────┬─────────┘
              │
      ┌───────┴────────┐
      │                │
      ▼                ▼
┌──────────┐    ┌──────────┐
│ Changed  │    │Same      │
│ Update   │    │Skip      │
└────┬─────┘    └────┬─────┘
     │               │
     └───────┬───────┘
             │
             ▼
        ┌─────────┐
        │  Repeat │
        └─────────┘
```

## Deployment Scenarios

### Scenario 1: Gradual Feature Rollout

```
Day 1: 10% of users
┌────────────────────────────────────┐
│ ████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │ 10%
└────────────────────────────────────┘

Day 3: 25% of users
┌────────────────────────────────────┐
│ ████████░░░░░░░░░░░░░░░░░░░░░░░░░░ │ 25%
└────────────────────────────────────┘

Day 7: 50% of users
┌────────────────────────────────────┐
│ ████████████████░░░░░░░░░░░░░░░░░░ │ 50%
└────────────────────────────────────┘

Day 14: 100% of users
┌────────────────────────────────────┐
│ ████████████████████████████████████│ 100%
└────────────────────────────────────┘
```

### Scenario 2: A/B Testing

```
Variant A (50%)          Variant B (50%)
┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │
│  Old Homepage   │      │  New Homepage   │
│                 │      │                 │
│  - Layout A     │      │  - Layout B     │
│  - Color A      │      │  - Color B      │
│  - CTA A        │      │  - CTA B        │
│                 │      │                 │
└─────────────────┘      └─────────────────┘
        │                        │
        └────────┬───────────────┘
                 │
                 ▼
         ┌───────────────┐
         │  Measure      │
         │  Conversion   │
         └───────────────┘
```

### Scenario 3: Kill Switch

```
Normal Operation
┌────────────────────────────────────┐
│  Feature: ENABLED                  │
│  ✓ Users can access feature        │
└────────────────────────────────────┘

Issue Detected
┌────────────────────────────────────┐
│  Admin: Disable flag via Redis     │
│  $ redis-cli SET feature_flags ... │
└────────────────────────────────────┘
                │
                ▼ (within 60s)
┌────────────────────────────────────┐
│  Feature: DISABLED                 │
│  ✗ Feature automatically disabled  │
│  ✓ No code deployment needed       │
└────────────────────────────────────┘
```
