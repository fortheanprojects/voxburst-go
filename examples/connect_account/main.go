// Example: Connect social media accounts and manage them
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fortheanprojects/voxburst-go/voxburst"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("VOXBURST_API_KEY")
	if apiKey == "" {
		log.Fatal("VOXBURST_API_KEY environment variable is required")
	}

	// Create client
	client := voxburst.NewClient(apiKey)

	ctx := context.Background()

	// Example 1: List all connected accounts
	fmt.Println("=== Connected Accounts ===")
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts.Data) == 0 {
		fmt.Println("No accounts connected yet.")
	} else {
		for _, acc := range accounts.Data {
			fmt.Printf("  - %s: %s (%s) - %s\n",
				acc.Platform,
				acc.Username,
				acc.DisplayName,
				acc.Status,
			)
		}
	}

	// Example 2: Filter accounts by platform
	fmt.Println("\n=== Twitter Accounts ===")
	twitterAccounts, err := client.Accounts.ListByPlatform(ctx, voxburst.PlatformTwitter)
	if err != nil {
		log.Fatalf("Failed to list Twitter accounts: %v", err)
	}

	if len(twitterAccounts) == 0 {
		fmt.Println("No Twitter accounts connected.")
	} else {
		for _, acc := range twitterAccounts {
			fmt.Printf("  - %s (%s)\n", acc.Username, acc.DisplayName)
		}
	}

	// Example 3: Initiate OAuth to connect a new account
	fmt.Println("\n=== Connecting New Account ===")
	oauth, err := client.Accounts.Connect(ctx, voxburst.PlatformTwitter, voxburst.ConnectParams{
		CallbackURL: "https://yourapp.com/oauth/callback",
	})
	if err != nil {
		log.Fatalf("Failed to initiate OAuth: %v", err)
	}

	fmt.Println("To connect a Twitter account:")
	fmt.Printf("1. Redirect user to: %s\n", oauth.AuthURL)
	fmt.Printf("2. Save state for verification: %s\n", oauth.State)
	fmt.Println("3. After user authorizes, complete with Callback()")

	// Example 4: Complete OAuth (after user returns from authorization)
	// This would be called in your callback handler
	/*
		callbackResult, err := client.Accounts.Callback(ctx, voxburst.PlatformTwitter, voxburst.CallbackParams{
			Code:  "authorization_code_from_callback",
			State: "state_from_original_request",
		})
		if err != nil {
			log.Fatalf("OAuth callback failed: %v", err)
		}
		fmt.Printf("Connected account: %s (%s)\n",
			callbackResult.Account.Username,
			callbackResult.Account.ID,
		)
	*/

	// Example 5: Get a specific account
	if len(accounts.Data) > 0 {
		fmt.Println("\n=== Account Details ===")
		acc, err := client.Accounts.Get(ctx, accounts.Data[0].ID)
		if err != nil {
			log.Fatalf("Failed to get account: %v", err)
		}
		fmt.Printf("Account: %s\n", acc.ID)
		fmt.Printf("  Platform: %s\n", acc.Platform)
		fmt.Printf("  Username: %s\n", acc.Username)
		fmt.Printf("  Display Name: %s\n", acc.DisplayName)
		fmt.Printf("  Type: %s\n", acc.AccountType)
		fmt.Printf("  Status: %s\n", acc.Status)
		fmt.Printf("  Connected: %s\n", acc.ConnectedAt.Format("2006-01-02"))
	}

	// Example 6: Find account by username
	fmt.Println("\n=== Find by Username ===")
	found, err := client.Accounts.GetByUsername(ctx, voxburst.PlatformTwitter, "@DarrenBuilds")
	if err != nil {
		if voxburst.IsNotFound(err) {
			fmt.Println("Account @DarrenBuilds not found on Twitter")
		} else {
			log.Fatalf("Error finding account: %v", err)
		}
	} else {
		fmt.Printf("Found: %s (ID: %s)\n", found.Username, found.ID)
	}

	// Example 7: Refresh account token
	if len(accounts.Data) > 0 {
		fmt.Println("\n=== Refreshing Token ===")
		if err := client.Accounts.Refresh(ctx, accounts.Data[0].ID); err != nil {
			// Token refresh might fail if the account needs re-authorization
			if apiErr, ok := voxburst.GetAPIError(err); ok {
				fmt.Printf("Refresh failed: %s - %s\n", apiErr.Code, apiErr.Message)
			} else {
				log.Fatalf("Failed to refresh token: %v", err)
			}
		} else {
			fmt.Println("Token refreshed successfully")
		}
	}

	// Example 8: Disconnect an account (be careful!)
	/*
		fmt.Println("\n=== Disconnecting Account ===")
		if err := client.Accounts.Disconnect(ctx, "acc_xyz789"); err != nil {
			if voxburst.IsNotFound(err) {
				fmt.Println("Account not found")
			} else {
				log.Fatalf("Failed to disconnect: %v", err)
			}
		} else {
			fmt.Println("Account disconnected")
		}
	*/

	// Example 9: Analytics for an account
	if len(accounts.Data) > 0 {
		fmt.Println("\n=== Account Analytics ===")
		analytics, err := client.Analytics.GetAccountMetrics(ctx, accounts.Data[0].ID, nil)
		if err != nil {
			log.Printf("Failed to get analytics: %v", err)
		} else {
			fmt.Printf("Account: %s\n", analytics.Username)
			fmt.Printf("  Followers: %d\n", analytics.Metrics.Followers)
			fmt.Printf("  Total Posts: %d\n", analytics.Metrics.TotalPosts)
			fmt.Printf("  Total Engagements: %d\n", analytics.Metrics.TotalEngagements)
			fmt.Printf("  Engagement Rate: %.2f%%\n", analytics.Metrics.EngagementRate)
		}
	}

	// Example 10: Analytics overview
	fmt.Println("\n=== Analytics Overview ===")
	overview, err := client.Analytics.GetOverview(ctx)
	if err != nil {
		log.Printf("Failed to get overview: %v", err)
	} else {
		fmt.Printf("Total Posts: %d\n", overview.Summary.TotalPosts)
		fmt.Printf("Total Impressions: %d\n", overview.Summary.TotalImpressions)
		fmt.Printf("Total Engagements: %d\n", overview.Summary.TotalEngagements)
		fmt.Printf("Avg Engagement Rate: %.2f%%\n", overview.Summary.AverageEngagementRate)

		fmt.Println("\nBy Platform:")
		for _, p := range overview.ByPlatform {
			fmt.Printf("  - %s: %d posts, %d impressions\n", p.Platform, p.Posts, p.Impressions)
		}
	}

	fmt.Println("\n✅ All examples completed!")
}
