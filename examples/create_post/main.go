// Example: Create and schedule a post with media
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	// List accounts first — you need account IDs to target specific connected accounts
	fmt.Println("=== Listing connected accounts ===")
	accounts, err := client.Accounts.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}
	for _, a := range accounts.Data {
		fmt.Printf("  %s  %s  (%s)\n", a.ID, a.Username, a.Platform)
	}
	// Use the IDs below — replace with real IDs from your workspace
	twitterAccountId := "acc_your_twitter_id"
	linkedinAccountId := "acc_your_linkedin_id"

	// Example 1: Create a simple draft post
	fmt.Println("\n=== Creating draft post ===")
	draft, err := client.Posts.CreateDraft(ctx,
		"Hello from the VoxBurst Go SDK! 🚀",
		[]string{twitterAccountId, linkedinAccountId},
	)
	if err != nil {
		log.Fatalf("Failed to create draft: %v", err)
	}
	fmt.Printf("Created draft: %s (status: %s)\n", draft.ID, draft.Status)

	// Example 2: Schedule a post for tomorrow
	fmt.Println("\n=== Scheduling post ===")
	scheduledTime := time.Now().Add(24 * time.Hour)
	scheduled, err := client.Posts.Schedule(ctx,
		"Scheduled post from Go SDK!",
		[]string{twitterAccountId},
		scheduledTime,
	)
	if err != nil {
		log.Fatalf("Failed to schedule post: %v", err)
	}
	fmt.Printf("Scheduled post: %s for %s\n", scheduled.ID, scheduled.ScheduledFor.Format(time.RFC3339))

	// Example 3: Create post with platform-specific content
	fmt.Println("\n=== Creating post with platform overrides ===")
	post, err := client.Posts.Create(ctx, voxburst.CreatePostParams{
		Content:    "Check out our new feature!",
		AccountIds: []string{twitterAccountId, linkedinAccountId},
		PlatformOverrides: map[voxburst.Platform]any{
			voxburst.PlatformTwitter: map[string]any{
				"content": "🎉 New feature alert! Check it out! #launch",
			},
			voxburst.PlatformLinkedIn: map[string]any{
				"content": "Excited to announce our latest feature that will help teams collaborate better. Here's what's new...",
			},
		},
		Metadata: map[string]any{
			"campaign": "feature-launch",
			"version":  "2.0",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create post: %v", err)
	}
	fmt.Printf("Created post with overrides: %s\n", post.ID)

	// Example 4: Upload media and create post with it
	fmt.Println("\n=== Creating post with media ===")

	// Upload a file (if you have one)
	// mediaID, err := client.Media.UploadFile(ctx, "/path/to/image.jpg")
	// if err != nil {
	//     log.Fatalf("Failed to upload media: %v", err)
	// }

	// Or upload raw bytes
	imageData := []byte{} // Your image bytes here
	if len(imageData) > 0 {
		mediaID, err := client.Media.Upload(ctx, "photo.jpg", "image/jpeg", imageData)
		if err != nil {
			log.Fatalf("Failed to upload media: %v", err)
		}

		instagramAccountId := "acc_your_instagram_id"
		postWithMedia, err := client.Posts.Create(ctx, voxburst.CreatePostParams{
			Content:     "Check out this photo!",
			AccountIds:  []string{twitterAccountId, instagramAccountId},
			ContentType: "IMAGE",
			Media:       []string{mediaID},
		})
		if err != nil {
			log.Fatalf("Failed to create post with media: %v", err)
		}
		fmt.Printf("Created post with media: %s\n", postWithMedia.ID)
	}

	// Example 5: List posts and paginate
	fmt.Println("\n=== Listing posts ===")
	listResp, err := client.Posts.List(ctx, &voxburst.ListPostsParams{
		Status: voxburst.PostStatusScheduled,
		Limit:  5,
	})
	if err != nil {
		log.Fatalf("Failed to list posts: %v", err)
	}

	fmt.Printf("Found %d scheduled posts\n", len(listResp.Data))
	for _, p := range listResp.Data {
		fmt.Printf("  - %s: %s\n", p.ID, truncate(p.Content, 50))
	}

	// Example 6: Iterate through all posts using iterator
	fmt.Println("\n=== Iterating all posts ===")
	iter := client.Posts.ListAll(ctx, &voxburst.ListPostsParams{
		Limit: 10, // Fetch 10 at a time
	})

	count := 0
	for iter.Next() {
		p := iter.Post()
		count++
		if count <= 3 {
			fmt.Printf("  - %s: %s\n", p.ID, truncate(p.Content, 50))
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("Error iterating posts: %v", err)
	}
	fmt.Printf("Total posts: %d\n", count)

	// Example 7: Update a post
	fmt.Println("\n=== Updating post ===")
	newContent := "Updated content!"
	updated, err := client.Posts.Update(ctx, draft.ID, voxburst.UpdatePostParams{
		Content: &newContent,
	})
	if err != nil {
		if voxburst.IsNotFound(err) {
			fmt.Println("Post not found (may have been deleted)")
		} else {
			log.Fatalf("Failed to update post: %v", err)
		}
	} else {
		fmt.Printf("Updated post: %s\n", updated.ID)
	}

	// Example 8: Publish a draft immediately
	fmt.Println("\n=== Publishing draft immediately ===")
	published, err := client.Posts.Publish(ctx, draft.ID)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Printf("Published: %s (status: %s)\n", published.ID, published.Status)

	// Example 9: Get post details
	fmt.Println("\n=== Getting post details ===")
	details, err := client.Posts.Get(ctx, published.ID)
	if err != nil {
		log.Fatalf("Failed to get post: %v", err)
	}
	fmt.Printf("Post %s:\n", details.ID)
	fmt.Printf("  Content: %s\n", truncate(details.Content, 50))
	fmt.Printf("  Status: %s\n", details.Status)
	for _, platform := range details.Platforms {
		fmt.Printf("  Platform: %s (%s)\n", platform.Platform, platform.Status)
	}

	// Example 10: Delete a post
	fmt.Println("\n=== Deleting scheduled post ===")
	if err := client.Posts.Delete(ctx, scheduled.ID); err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}
	fmt.Println("Deleted scheduled post")

	fmt.Println("\n✅ All examples completed!")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
