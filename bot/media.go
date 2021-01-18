package bot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func getVideosList(keyWord, token string) ([]Video, error) {
	service, err := youtube.NewService(context.TODO(), option.WithAPIKey(token))
	if err != nil {
		return nil, fmt.Errorf("Error during creating new service: %w", err)
	}

	request := service.Search.List("id, snippet").
		Q(keyWord).
		MaxResults(5).
		Order("relevance").
		Type("video")

	listVideos, err := request.Do()
	if err != nil {
		return nil, fmt.Errorf("Youtube request error: %w", err)
	}

	var videos []Video

	for _, item := range listVideos.Items {
		video := Video{}
		video.Title = html.UnescapeString(item.Snippet.Title)
		video.ID = item.Id.VideoId

		videos = append(videos, video)
	}

	return videos, nil
}

func downloadAudioByLink(link, downloadPath string) (string, error) {
	arg := fmt.Sprintf(`youtube-dl --get-filename -o '%%(title)s.mp3' '%s'`, link)
	cmd := exec.Command("sh", "-c", arg)

	var stdBuff bytes.Buffer
	cmd.Stderr = &stdBuff

	fileName, err := cmd.Output()
	if err != nil {
		return "", errors.New(err.Error() + ":" + stdBuff.String())
	}

	arg = fmt.Sprintf(`youtube-dl -o'%s/%%(title)s.%%(ext)s' -k -x --audio-quality 0 --retries 5 --no-part --audio-format mp3 -f 'bestaudio[filesize<10M]' '%s'`,
		downloadPath, link)
	cmd = exec.Command("sh", "-c", arg)

	mw := io.MultiWriter(os.Stdout, &stdBuff)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Run()
	if err != nil {
		return "", errors.New(err.Error() + ":" + stdBuff.String())
	}

	return strings.TrimSpace(string(fileName)), nil
}
